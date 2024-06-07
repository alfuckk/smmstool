package main

import (
	"fmt"

	"github.com/5asp/smmstool"
)

func main() {
	smmstools := smmstool.NewSmmsTool()
	// 展示菜单
	for {
		fmt.Println("请选择操作：")
		fmt.Println("1. 登录")
		fmt.Println("2. 上传")
		fmt.Println("3. 查看额度")
		fmt.Println("4. 选择区域")
		fmt.Println("5. 退出")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			smmstools.Login()
		case 2:
			smmstools.Upload()
		case 3:
			smmstools.SelectMode()
		case 4:
			fmt.Println("退出程序")
			return
		default:
			fmt.Println("无效的选择，请重新选择")
		}
	}
}
