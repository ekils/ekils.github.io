package cronjobs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Script(companies []string) error {

	fmt.Println("腳本: 0")
	// 从环境变量中获取GitHub Token
	gt := "abcd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"
	githubToken := gt[3:]
	if githubToken == "" {
		log.Fatal("GT is not set")
	}
	fmt.Printf("腳本: 0-1(githubToken): %v \n", githubToken)
	// 设置Git配置，使用GitHub Token进行身份验证
	gitConfigCmd := exec.Command("git", "config", "--global", "credential.helper", "!f() { echo username=x-access-token; echo password=$GT; }; f")
	gitConfigCmd.Env = append(os.Environ(), fmt.Sprintf("GT=%s", githubToken))
	if err := gitConfigCmd.Run(); err != nil {
		log.Fatalf("Failed to configure git: %v", err)
	}
	fmt.Println("腳本: 0-2")
	combinedCmd := exec.Command("sh", "-c", `pwd`)
	output, err := combinedCmd.Output()
	if err != nil {
		fmt.Println("執行命令時發生錯誤:", err)
		return err
	}
	fmt.Println("腳本: 1")
	fmt.Println("目前所在的工作目錄1:", string(output))
	// 1
	cmd := `
	rm -rf ./plot/*.html;
	cp ./html/*.html ./plot/;
	cp temp ./plot/index.html;
	cd plot;
	sed -i '$ d' index.html
	`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤1:", err)
		return err
	}
	//2
	fmt.Println("腳本: 2")
	fmt.Println("目前所在的工作目錄2:", string(output))
	for _, company := range companies {

		cmd := fmt.Sprintf(`
			cd plot;
			echo " <a href="https://ekils.github.io/plot/PE_Trend_%s.html" target="_blank">%s</a><br><br>"  >> index.html`, company, company)
		combinedCmd := exec.Command("sh", "-c", cmd)
		if err := combinedCmd.Run(); err != nil {
			fmt.Println("執行命令時發生錯誤2:", err)
			return err
		}
	}
	//3
	fmt.Println("腳本: 3")
	fmt.Println("目前所在的工作目錄3:", string(output))
	cmd = `
		cd plot;
		echo " </body></html>"  >> index.html
		current_date=$(date);`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤3:", err)
		return err
	}

	fmt.Println("腳本: 3-1")
	cmd = `
		git config --global user.email "bobobo746@hotmail.com";
		git config --global user.name "ekils";`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤3-1:", err)
		return err
	}
	fmt.Println("腳本: 3-2")
	cmd = `
	   git status --porcelain; `
	combinedCmd = exec.Command("sh", "-c", cmd)
	var out bytes.Buffer
	combinedCmd.Stdout = &out
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤3-2:", err)
		return err
	}
	gitStatusOutput := out.String()
	if strings.TrimSpace(gitStatusOutput) == "" {
		fmt.Println("沒有檔案更新, 不用推 git")
		return nil
	} else {
		fmt.Printf("腳本: 3-3(有更新檔案): %v", gitStatusOutput)
		cmd = `
			git add .;
			git commit -m "Modify Version: $current_date";`
		combinedCmd = exec.Command("sh", "-c", cmd)
		if err := combinedCmd.Run(); err != nil {
			fmt.Println("執行命令時發生錯誤3-3:", err)
			return err
		}
		// 推送到GitHub

		fmt.Println("腳本: 4")
		fmt.Println("目前所在的工作目錄4:", string(output))
		cmd = `export GT=` + githubToken + ` git push --set-upstream https://github.com/ekils/ekils.github.io.git main; `
		combinedCmd = exec.Command("sh", "-c", cmd)

		// combinedCmd.Env = append(os.Environ(), fmt.Sprintf("GT=%s", githubToken))
		if err := combinedCmd.Run(); err != nil {
			fmt.Println("執行命令時發生錯誤4:", err)
			return err
		} else {
			fmt.Println("shell script done ..")
			return nil
		}
	}
}

//api key: rnd_TUXka6fC14B8euPEBVXpN3UhFif5
