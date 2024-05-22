package cronjobs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Script(companies []string) error {

	fmt.Println("腳本: 前置作業 git init ")
	// 初始化Git仓库（如果尚未初始化）
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		if err := exec.Command("git", "init").Run(); err != nil {
			fmt.Printf("Failed to initialize git repository: %v", err)
		}
	}

	fmt.Println("腳本: 前置作業 email, name setting .....")
	cmd := `
	git config --global user.email "bobobo746@hotmail.com";
	git config --global user.name "ekils"
	`

	combinedCmd := exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("email, name setting 發生錯誤:", err)
		return err
	}

	fmt.Println("腳本: 前置作業 credential.helper ")
	cmd = `
	git config --global credential.helper '!f() { echo username=x-access-token; echo password=$GT; }; f'`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("credential.helper 發生錯誤 :", err)
		return err
	}

	fmt.Println("腳本:設置遠程git hub URL")
	remoteURL := "https://github.com/ekils/ekils.github.io.git`"
	if err := exec.Command("git", "remote", "add", "origin", remoteURL).Run(); err != nil && !isRemoteAlreadyExists(err) {
		fmt.Printf("Failed to set remote repository: %v", err)
		return err
	}

	fmt.Println("腳本: checkout")
	cmd = `
	git checkout main`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("checkout 發生錯誤 :", err)
		return err
	}

	fmt.Println("腳本: 1")
	outCmd := exec.Command("sh", "-c", `pwd`)
	output, err := outCmd.Output()
	if err != nil {
		fmt.Println("執行命令時發生錯誤:", err)
		return err
	}

	fmt.Println("目前所在的工作目錄:", string(output))
	// 1
	cmd = `
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
	cmd = `git ls-remote https://github.com/ekils/ekils.github.io.git`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤3-1 : Git 登入失敗:", err)
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
		fmt.Printf("腳本: 3-4(有更新檔案): %v\n", gitStatusOutput)
		cmd = `
			git add .;
			git commit -m "Modify Version: $current_date";`
		combinedCmd = exec.Command("sh", "-c", cmd)
		if err := combinedCmd.Run(); err != nil {
			fmt.Println("執行命令時發生錯誤3-4:", err)
			return err
		}
		// 推送到GitHub
		fmt.Println("腳本: 4")
		cmd = ` git push --set-upstream https://github.com/ekils/ekils.github.io.git main; `
		combinedCmd = exec.Command("sh", "-c", cmd)
		if err := combinedCmd.Run(); err != nil {
			fmt.Println("執行命令時發生錯誤4:", err)
			return err
		} else {
			fmt.Println("shell script done ..")
			return nil
		}
	}
}

// isRemoteAlreadyExists 检查远程仓库是否已经存在
func isRemoteAlreadyExists(err error) bool {
	return err != nil && err.Error() == "exit status 128"
}

// 備註:
// git config --global --list
// git config --global credential.helper store
// git config --global --unset credential.helper

// credential.helper=!f() { echo username=x-access-token; echo password=$GT; }; f

// git config --global credential.helper '!f() { echo username=x-access-token; echo password=$GT; }; f'

// echo "[user]" > /opt/render/.gitconfig ;
// echo "    email = bobobo746@hotmail.com" >> /opt/render/.gitconfig;
// echo "    name = ekils" >> /opt/render/.gitconfig;
// echo "[credential]" >> /opt/render/.gitconfig;
// echo '    helper = "!f() { echo username=x-access-token; echo password=$GT; }; f"' >> /opt/render/.gitconfig

// render@srv-cp64qq021fec738b7gdg-774f9f8b79-4gtcv:~/project/go/src/github.com/ekils/ekils.github.io$ pwd
// /opt/render/project/go/src/github.com/ekils/ekils.github.io
