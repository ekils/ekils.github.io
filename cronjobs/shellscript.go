package cronjobs

import (
	"fmt"
	"os/exec"
)

func Script(companies []string) error {

	combinedCmd := exec.Command("sh", "-c", `pwd`)
	output, err := combinedCmd.Output()
	if err != nil {
		fmt.Println("執行命令時發生錯誤:", err)
		return err
	}

	fmt.Println("目前所在的工作目錄:", string(output))
	// 1
	// cmd := `
	// rm -rf ./ekils.github.io/plot/*.html;
	// cp ./html/*.html ./ekils.github.io/plot/;
	// cd ./ekils.github.io;
	// cp temp ./plot/index.html;
	// cd plot;
	// sed -i '$ d' index.html
	// `
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
	cmd = `
		cd plot;
		echo " </body></html>"  >> index.html
		current_date=$(date);
		git config --global user.email "bobobo746@hotmail.com";
		git config --global user.name "ekils";
		git add .;
		git commit -m "Modify Version: $current_date";
		git push;
	`
	combinedCmd = exec.Command("sh", "-c", cmd)
	if err := combinedCmd.Run(); err != nil {
		fmt.Println("執行命令時發生錯誤3:", err)
		return err
	}
	fmt.Println("shell script done ..")
	return nil

}
