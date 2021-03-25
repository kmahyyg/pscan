package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type TCPScanner struct {
	scanNets    []string
	scanPorts   []int
	scanTimeout time.Duration
	scanThreads int
}

var (
	portWeb            = []int{443, 80, 8080, 7001, 7002, 9060, 9080, 9443, 8443, 3000, 9000, 9090}
	portRemote         = []int{22, 5938, 5985, 3389, 1080, 5800, 5900}
	portFileTrans      = []int{21, 23, 139, 445, 135, 2121, 2049, 3690}
	portDatabase       = []int{3306, 27017, 1433, 1521, 61616, 6379, 9200, 15672, 5432}
	portVirtualization = []int{902, 903, 2375, 5000}
	portSpecial        = []int{8090, 8009, 4430, 7012, 8088, 18080, 28080}
)

func connectTCP(targets chan string, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	for target := range targets {
		if len(target) < 2 {
			return
		}
		conn, err := net.DialTimeout("tcp", target, timeout)
		if err != nil || conn == nil {
			if strings.Contains(err.Error(), "too many open files") {
				time.Sleep(time.Duration(flagTimeout) * time.Second)
			} else {
				fmt.Fprintf(os.Stderr, "%s CLOSED! \n", target)
			}
		} else {
			defer conn.Close()
			log.Printf("%s OPEN! \n", target)
		}
	}
}

func (pscanner *TCPScanner) Scan() {
	var wg = sync.WaitGroup{}
	var targets = make(chan string, runtime.NumCPU()*4)
	for t := 0; t < pscanner.scanThreads; t++ {
		wg.Add(1)
		go connectTCP(targets, pscanner.scanTimeout, &wg)
	}
	for _, vport := range pscanner.scanPorts {
		for _, vip := range pscanner.scanNets {
			target := fmt.Sprintf("%s:%d", vip, vport)
			targets <- target
		}
	}
	close(targets)
	wg.Wait()
}
