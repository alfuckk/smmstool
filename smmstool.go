package smmstool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

type SmmsTool struct {
	config   *ini.File
	confPath string
}

func NewSmmsTool() *SmmsTool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
		return nil
	}
	targetDir := filepath.Join(homeDir, ".smms")
	cfg := &ini.File{}
	if !FileExists(targetDir) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.Fatalf("error mkdir directory: %v", err)
			return nil
		}
	}

	configDir := filepath.Join(targetDir, "config.ini")
	if !FileExists(configDir) {
		fmt.Println(configDir)
		myfile, e := os.Create(configDir)
		if e != nil {
			log.Fatalf("error create ini: %v", err)
		}
		myfile.Close()
	}
	cfg, err = ini.Load(configDir)
	if err != nil {
		log.Fatalf("error read ini: %v", err)
	}
	if mode, err := cfg.Section("smms").Key("mode").Int(); err != nil {
		if mode == 0 {
			cfg.Section("smms").Key("mode").SetValue("1")
		}
	}
	if err = cfg.SaveTo(configDir); err != nil {
		log.Fatalf("Fail to save file: %v", err)
	}
	return &SmmsTool{
		config:   cfg,
		confPath: configDir,
	}
}

func (st *SmmsTool) getMode() (addr string) {
	mode, err := st.config.Section("smms").Key("mode").Int()
	if err != nil {
		log.Fatalf("smms mode error: %v", err)
	}

	if mode == 1 {
		addr = "https://sm.ms/api/v2/"
	} else {
		addr = "https://smms.app/api/v2/"
	}
	return addr
}

func (st *SmmsTool) getToken() (token string) {
	token = st.config.Section("smms").Key("token").String()
	if token == "" {
		log.Fatalf("smms token not exist.")
	}
	return
}

func (st *SmmsTool) Login() {
	fmt.Println("请输入用户名:")
	var username string
	fmt.Scanln(&username)

	fmt.Println("请输入密码:")
	var password string
	fmt.Scanln(&password)

	st.config.Section("smms").Key("username").SetValue(username)
	st.config.Section("smms").Key("password").SetValue(password)

	formData := &URLValuesWrapper{
		data: url.Values{
			"username": {username},
			"password": {password},
		},
	}

	// 设置总超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	addr := st.getMode()
	// 发送请求并重试
	retries := 30
	resp, err := sendRequestWithRetry(ctx, addr+"token", "", formData, retries)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	var data LoginData
	_, err = processResponse(string(bodyBytes), &data)
	if err != nil {
		fmt.Println("Error processing response:", err)
		return
	}
	st.config.Section("smms").Key("token").SetValue(data.Token)
	st.config.SaveTo(st.confPath)

	fmt.Println("登录信息已保存")
	fmt.Printf("Token Authorization: %s \n", data.Token)
}

func (st *SmmsTool) Upload() {
	fmt.Println("拖拽文件路径到这:")
	var smfilePath string
	fmt.Scanln(&smfilePath)

	file, err := os.Open(smfilePath)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return
	}
	defer file.Close()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	part, err := writer.CreateFormFile("smfile", filepath.Base(smfilePath))
	if err != nil {
		log.Fatalf("create form file error: %v", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatalf("copy file error: %v", err)
		return
	}
	contentType := writer.FormDataContentType()

	err = writer.Close()
	if err != nil {
		log.Fatalf("close writer error: %v", err)
		return
	}

	addr := st.getMode()
	token := st.getToken()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	retries := 3

	formData := &BufferWrapper{
		data:        payload,
		contentType: contentType,
	}

	resp, err := sendRequestWithRetry(ctx, addr+"upload", token, formData, retries)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	var data UploadData
	code, err := processResponse(string(bodyBytes), &data)
	if err != nil {
		fmt.Println("Error processing response:", err)
		return
	}
	if code == "image_repeated" {
		fmt.Println("注意不能重复上传.")
		return
	}
	fmt.Printf("源图片: %s \n", data.Filename)
	fmt.Printf("源图片宽: %d \n", data.Width)
	fmt.Printf("源图片高: %d \n", data.Height)
	fmt.Printf("源图片大小: %d \n", data.Size)
	fmt.Printf("远程URL: %s \n", data.URL)
}

func (st *SmmsTool) SelectMode() {
	fmt.Println("请选择模式：1. 正常，2. cn")
	var mode int
	fmt.Scanln(&mode)

	if mode != 1 && mode != 2 {
		fmt.Println("无效的选择，请重新选择")
		return
	}

	st.config.Section("smms").Key("mode").SetValue(strconv.Itoa(mode))
	st.config.SaveTo(st.confPath)

	fmt.Println("模式已更新")
}
