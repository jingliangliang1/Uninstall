package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"golang.org/x/sys/windows/registry"
	"strings"
)

// 删除目录下的所有文件
func deleteFilesRecursively(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("无法删除文件夹 %s: %v", path, err)
	}
	fmt.Printf("删除文件夹：%s\n", path)
	return nil
}

// 清理注册表中的相关项
func deleteRegistryKey(keyPath string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("无法打开注册表项: %v", err)
	}
	defer key.Close()

	// 删除整个注册表项
	err = registry.DeleteKey(registry.CURRENT_USER, keyPath)
	if err != nil {
		return fmt.Errorf("删除注册表项失败: %v", err)
	}

	fmt.Printf("删除注册表项：%s\n", keyPath)
	return nil
}

func main() {
	// 打开注册表路径，访问已安装的软件
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.READ)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	// 获取所有子键（每个子键对应一个已安装的软件）
	names, err := key.ReadSubKeyNames(0)
	if err != nil {
		log.Fatal(err)
	}

	// 显示软件列表
	fmt.Println("已安装的软件：")
	softwareList := make(map[int]string)
	for i, name := range names {
		softwareKey, err := registry.OpenKey(key, name, registry.READ)
		if err != nil {
			continue
		}
		defer softwareKey.Close()

		// 获取软件的显示名称和卸载命令
		displayName, _, err := softwareKey.GetStringValue("DisplayName")
		if err == nil {
			softwareList[i+1] = displayName
			fmt.Printf("%d. %s\n", i+1, displayName)
		}
	}

	// 让用户选择一个软件进行卸载
	var choice int
	fmt.Print("请输入要卸载的软件编号：")
	fmt.Scanf("%d", &choice)

	if softwareName, exists := softwareList[choice]; exists {
		// 获取软件的卸载命令
		softwareKey, err := registry.OpenKey(key, names[choice-1], registry.READ)
		if err != nil {
			log.Fatal(err)
		}
		defer softwareKey.Close()

		uninstallString, _, err := softwareKey.GetStringValue("UninstallString")
		if err != nil {
			log.Fatal("未找到卸载命令:", err)
		}

		// 执行卸载命令
		fmt.Printf("卸载命令: %s\n", uninstallString)
		cmd := exec.Command("cmd", "/C", uninstallString)
		err = cmd.Run()
		if err != nil {
			log.Fatal("卸载失败:", err)
		} else {
			fmt.Println("卸载成功！")
		}

		// 删除软件的残余文件
		installLocation, _, err := softwareKey.GetStringValue("InstallLocation")
		if err == nil && installLocation != "" {
			err := deleteFilesRecursively(installLocation)
			if err != nil {
				log.Println("删除残余文件失败:", err)
			}
		}

		// 清理注册表项
		err = deleteRegistryKey(fmt.Sprintf("SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\%s", names[choice-1]))
		if err != nil {
			log.Println("清理注册表项失败:", err)
		}

		fmt.Println("清理完成！")
	} else {
		fmt.Println("无效选择！")
	}
}
