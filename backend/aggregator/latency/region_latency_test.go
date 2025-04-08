package latency

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-ping/ping"
)

type Target struct {
    IP     string
    Region string
}

type PingResult struct {
    SourceHost   string
    SourceRegion string
    TargetIP     string
    TargetRegion string
    Timestamp    string
    MinLatency   float64
    AvgLatency   float64
    MaxLatency   float64
    StdDev       float64
    PacketLoss   float64
}

func TestRegionLatency(t *testing.T) {
    sourceRegion := "ap-northeast-2" // 현재 사용 중인 리전으로 변경
    sourceHostname, _ := os.Hostname()

    targets := []Target{
        {"13.231.217.241", "ap-northeast-1"},
        {"52.77.229.124", "ap-southeast-1"},
        {"3.80.156.106", "us-east-1"},
    }

    resultsFile := fmt.Sprintf("aws_regionlatency%s_%s.csv", 
        sourceHostname, time.Now().Format("20060102_150405"))

    // CSV 파일 생성
    file, err := os.Create(resultsFile)
    if err != nil {
        fmt.Printf("파일 생성 오류: %v\n", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // CSV 헤더 작성
    writer.Write([]string{
        "Source", "SourceRegion", "Target", "TargetRegion", "Timestamp",
        "MinLatency(ms)", "AvgLatency(ms)", "MaxLatency(ms)", "StdDev(ms)", "PacketLoss(%)",
    })

    var wg sync.WaitGroup
    resultsChan := make(chan PingResult, len(targets))

    fmt.Println("== 리전 간 latency 측정 시작 ==")

    // 각 대상에 대해 병렬로 ping 실행
    for _, target := range targets {
        wg.Add(1)
        go func(t Target) {
            defer wg.Done()

            fmt.Printf("리전 %s에서 리전 %s(%s)으로 latency 측정 중...\n", 
                sourceRegion, t.Region, t.IP)

            pinger, err := ping.NewPinger(t.IP)
            if err != nil {
                fmt.Printf("Pinger 생성 오류 (%s): %v\n", t.IP, err)
                return
            }

            // 일부 시스템에서는 권한 설정이 필요할 수 있음
            pinger.SetPrivileged(true)

            pinger.Count = 5
            pinger.Timeout = 2 * time.Second

            err = pinger.Run()
            if err != nil {
                fmt.Printf("Ping 실행 오류 (%s): %v\n", t.IP, err)
                return
            }

            stats := pinger.Statistics()

            result := PingResult{
                SourceHost:   sourceHostname,
                SourceRegion: sourceRegion,
                TargetIP:     t.IP,
                TargetRegion: t.Region,
                Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
                MinLatency:   float64(stats.MinRtt) / float64(time.Millisecond),
                AvgLatency:   float64(stats.AvgRtt) / float64(time.Millisecond),
                MaxLatency:   float64(stats.MaxRtt) / float64(time.Millisecond),
                StdDev:       float64(stats.StdDevRtt) / float64(time.Millisecond),
                PacketLoss:   stats.PacketLoss,
            }

            resultsChan <- result

            fmt.Printf("결과: 최소=%.2fms, 평균=%.2fms, 최대=%.2fms, 패킷손실=%.1f%%\n",
                result.MinLatency, result.AvgLatency, result.MaxLatency, result.PacketLoss)
            fmt.Println("----------------------------------------")
        }(target)
    }

    // 모든 ping 완료 대기
    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    // 결과 수집 및 CSV에 저장
    for result := range resultsChan {
        writer.Write([]string{
            result.SourceHost,
            result.SourceRegion,
            result.TargetIP,
            result.TargetRegion,
            result.Timestamp,
            fmt.Sprintf("%.2f", result.MinLatency),
            fmt.Sprintf("%.2f", result.AvgLatency),
            fmt.Sprintf("%.2f", result.MaxLatency),
            fmt.Sprintf("%.2f", result.StdDev),
            fmt.Sprintf("%.1f", result.PacketLoss),
        })
        writer.Flush()
    }

    fmt.Printf("\n모든 측정 완료. 결과는 %s 파일에 저장되었습니다.\n", resultsFile)

    // 결과 요약 출력
    fmt.Println("\n== 리전 간 latency 측정 결과 요약 ==")
    fmt.Printf("출발지 리전: %s\n\n", sourceRegion)
    fmt.Println("대상 리전          | 평균 Latency (ms) | 패킷 손실 (%)")
    fmt.Println("--------------------|-------------------|---------------")

    // 요약 데이터 다시 읽기
    file, _ = os.Open(resultsFile)
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Read() // 헤더 건너뛰기

    for {
        record, err := reader.Read()
        if err != nil {
            break
        }

        targetRegion := record[3]
        avgLatency := record[6]
        packetLoss := record[9]

        fmt.Printf("%-20s | %-19s | %s\n", targetRegion, avgLatency, packetLoss)
    }
}