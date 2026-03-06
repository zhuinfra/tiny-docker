package container

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	RUNNING             = "running"
	STOP                = "stopped"
	Exit                = "exited"
	DefaultInfoLocation = "/var/run/tiny-docker/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
	Volume      string `json:"volume"`
}

func randStringBytes(n int) string {
	letterBytes := "1234567890" // 其实是 0-9，顺序不影响随机结果
	b := make([]byte, n)
	for i := range b {
		// 直接用 rand.Intn，全局生成器已自动初始化
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateId() string {
	return randStringBytes(10)
}

// 记录容器信息
func RecordContainerInfo(containerID string, containerPID int, commandArray []string, containerName string, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, " ")
	containInfo := ContainerInfo{
		Pid:         strconv.Itoa(containerPID),
		Id:          containerID,
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Volume:      volume,
	}

	// 容器信息保存成json
	jsonBytes, err := json.Marshal(containInfo)
	if err != nil {
		slog.Error("json.Marshal error")
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	if err := os.MkdirAll(dirUrl, 0777); err != nil {
		slog.Error("Mkdir error.")
		return "", err
	}
	filePath := dirUrl + ConfigName
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Create file error.")
		return "", err
	}
	defer file.Close()
	if _, err := io.WriteString(file, jsonStr); err != nil {
		slog.Error("Write string error.")
		return "", err
	}
	return filePath, nil
}

func DeleteContainerInfo(containerID string) error {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerID)
	return os.RemoveAll(dirUrl)
}
