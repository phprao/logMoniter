package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	TimeLocal                    time.Time
	BytesSent                    int
	Path, Method, Scheme, Status string
	UpstreamTime, RequestTime    float64
}

func main() {
	// 接受命令行参数 go run main.go -path aaa.log -influxDBDsn sdc
	var path, influxDBDsn string
	flag.StringVar(&path, "path", "./access.log", "read file path")
	flag.StringVar(&influxDBDsn, "influxDBDsn", "", "influxDBDsn info")
	flag.Parse()

	lp := &LogProcess{
		rchan:        make(chan []byte, 200),   // 使用 buffer channel 优化性能
		wchan:        make(chan *Message, 200), // 使用 buffer channel 优化性能
		readHandler:  &FileReader{path: path},  // 增加程序扩展性
		WriteHandler: &DbWriter{influxDBDsn: influxDBDsn}, // 增加程序扩展性
	}

	// 读取操作
	go lp.ReadDataModule()

	// 多开几个运算处理协程，Go语言保证了同一时刻只有一个goroutine能访问channel里的数据，从语言层面支持并发处理
	for pNum := 1; pNum <= 2; pNum++ {
		// 处理程序
		go lp.Process()
	}

	// 多开几个记录模块协程
	for wNum := 1; wNum <= 2; wNum++ {
		// 分析结果记录
		go lp.WriteDataModule()
	}

	// 程序运行状况
	m := &Monitor{
		startTime: time.Now(),
		data:      SystemInfo{},
	}
	m.start(lp)
}

type LogProcess struct {
	rchan        chan []byte
	wchan        chan *Message
	readHandler  Reader
	WriteHandler Writer
}

func (c *LogProcess) ReadDataModule() {
	c.readHandler.ReadData(c.rchan)
}

/**
示例
127.0.0.1 - - [21/Dec/2015:20:22:14 +0800] http "GET /phpinfo.php HTTP/1.1" 200 12704 "-" "KeepAliveClient" "-" 1.005 1.854

正则
([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)

*/

func (c *LogProcess) Process() {
	// 使用正则匹配
	r := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)

	// 设置时区
	location, _ := time.LoadLocation("Asia/Shanghai")

	for v := range c.rchan {
		res := r.FindStringSubmatch(string(v))
		if len(res) != 14 {
			TypeMonitorChan <- TypeErrNum
			log.Println("匹配异常：", len(res))
			continue
		}

		t, err := time.ParseInLocation("02/Jan/2006:15:04:05 +0800", res[4], location)
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("时间解析错误：", res[4])
			continue
		}

		// 字符转int  12704
		bytes, _ := strconv.Atoi(res[8])

		// GET /phpinfo.php HTTP/1.1
		s1 := strings.Split(res[6], " ")
		if len(s1) != 3 {
			TypeMonitorChan <- TypeErrNum
			log.Println("path解析错误：", res[6])
			continue
		}

		u, err := url.Parse(s1[1])
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("path解析错误：", s1[1])
			continue
		}

		UpstreamTime, _ := strconv.ParseFloat(res[12], 64)
		RequestTime, _ := strconv.ParseFloat(res[13], 64)

		msg := &Message{
			TimeLocal:    t,
			BytesSent:    bytes,
			Path:         u.Path,
			Method:       s1[0],
			Scheme:       res[5],
			Status:       res[7],
			UpstreamTime: UpstreamTime,
			RequestTime:  RequestTime,
		}

		// 将分析结果写入通道
		c.wchan <- msg
	}
}

func (c *LogProcess) WriteDataModule() {
	c.WriteHandler.WriteData(c.wchan)
}

// ------------- 读取模块 -------------

type Reader interface {
	ReadData(rchan chan []byte)
}

type FileReader struct {
	path string // 读取文件路径
}

func (f *FileReader) ReadData(rchan chan []byte) {
	file, err := os.Open(f.path)
	if err != nil {
		panic(fmt.Sprintf("open file error: %s", err.Error()))
	}
	defer file.Close()

	// 从文件末尾开始逐行读取，防止重复统计之前的记录
	file.Seek(0, 2) // 将指针定位到文件末尾
	newFile := bufio.NewReader(file)

	for {
		// 读取一行，以\n符作为分割
		line, err := newFile.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond) // 等待500毫秒，考虑到日志文件是动态写入的
			log.Println("reach file end...")
			continue
		} else if err != nil {
			panic(fmt.Sprintf("read file error: %s", err.Error()))
		}
		rchan <- line[:len(line)-1] // 去掉换行符，写入通道

		TypeMonitorChan <- TypeHandleLine
	}
}

// ------------- 写入模块 -------------

type Writer interface {
	WriteData(wchan chan *Message)
}

type DbWriter struct {
	influxDBDsn string // 数据库连接
}

func (f *DbWriter) WriteData(wchan chan *Message) {
	// 模拟写入操作
	for v := range wchan {
		fmt.Println(v)
	}
}

// ------------- 程序的健康状况 -------------

const (
	TypeHandleLine = 0 // 读取标记
	TypeErrNum     = 1 // 错误标记
	TickerSecond   = 1 // tps统计间隔
)

var TypeMonitorChan = make(chan int, 200)

type SystemInfo struct {
	HandleLine   int     `json:"handleLine"`   // 总处理行数
	Tps          float64 `json:"tps"`          // 系统吞吐量
	ReadChanLen  int     `json:"readChanLen"`  // read channel 待处理长度
	WriteChanLen int     `json:"writeChanLen"` // write channel 待处理长度
	RunTime      string  `json:"runTime"`      // 运行总时长
	ErrNum       int     `json:"errNum"`       // 错误数
}

type Monitor struct {
	data      SystemInfo
	startTime time.Time
	tpsSli    []int // 记录Tps
}

func (m *Monitor) start(lp *LogProcess) {
	// 消费 TypeMonitorChan
	go func() {
		for n := range TypeMonitorChan {
			switch n {
			case TypeHandleLine:
				m.data.HandleLine++
			case TypeErrNum:
				m.data.ErrNum++
			}
		}
	}()

	// 使用定时器统计Tps
	ticker := time.NewTicker(time.Second * TickerSecond)
	go func() {
		for {
			<-ticker.C
			m.tpsSli = append(m.tpsSli, m.data.HandleLine)
			// 秩序保留最新的2个即可
			if len(m.tpsSli) > 2 {
				m.tpsSli = m.tpsSli[1:]
			}
		}
	}()

	// 提供http服务
	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rchan)
		m.data.WriteChanLen = len(lp.wchan)
		// 计算Tps
		if len(m.tpsSli) >= 2 {
			m.data.Tps = float64((m.tpsSli[1] - m.tpsSli[0]) / TickerSecond)
		}

		ret, _ := json.MarshalIndent(m.data, "", "\t")

		io.WriteString(writer, string(ret))
	})

	// 监听端口 9193
	http.ListenAndServe(":9193", nil)
}


