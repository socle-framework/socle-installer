package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/danielkeho/crypto/pkg/random"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var (
	arch                 string
	db                   string
	http                 string
	render               string
	with                 []string
	template             string
	force                bool
	appURL, skeletonName string
)

var newCmd = &cobra.Command{
	Use:   "new <project_name>",
	Short: "Create a new Socle project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		fmt.Println("🚀 Creating project:", projectName)
		fmt.Println("📦 Architecture:", arch)
		fmt.Println("🛢️ Database:", db)
		fmt.Println("🌐 HTTP Framework:", http)
		fmt.Println("🖼️ Template Engine:", render)
		fmt.Println("📦 Modules:", strings.Join(with, ", "))
		fmt.Println("📁 Force overwrite:", force)

		// Exemple : vérifie si le dossier existe
		if _, err := os.Stat(projectName); err == nil && !force {
			fmt.Println("❌ Folder already exists. Use --force to overwrite.")
			return
		}
		doNew(projectName)
	},
}

func init() {
	newCmd.Flags().StringVarP(&arch, "arch", "a", "ddd", "Architecture (ddd, layered, microservice, minimal)")
	newCmd.Flags().StringVar(&db, "db", "sqlite", "Database engine")
	newCmd.Flags().StringVar(&http, "http", "chi", "HTTP framework")
	newCmd.Flags().StringVar(&render, "render", "templ", "Template engine")
	newCmd.Flags().StringSliceVar(&with, "with", []string{}, "Modules to include (comma-separated)")
	newCmd.Flags().StringVar(&template, "template", "", "Custom Git template")
	newCmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing folder")

	rootCmd.AddCommand(newCmd)
}

func doNew(appName string) {
	appName = strings.ToLower(appName)
	appURL = appName

	// sanitize the application name (convert url to single word)
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded) - 1)]
	}

	log.Println("App name is", appName)

	// git clone the skeleton application
	color.Green("\tCloning repository...")
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "https://gitlab.com/socle-framework/starter.git",
		Progress: os.Stdout,
		Depth:    1,
	})
	//err := doSkeleton(skeletonName, appName)

	if err != nil {
		exitGracefully(err)
	}

	// remove .git directory
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		exitGracefully(err)
	}

	// create a ready to go .env file
	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		exitGracefully(err)
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", random.RandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		exitGracefully(err)
	}

	// create a makefile
	if runtime.GOOS == "windows" {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.windows", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	} else {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.mac", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	}
	_ = os.Remove("./" + appName + "/Makefile.mac")
	_ = os.Remove("./" + appName + "/Makefile.windows")

	// update the go.mod file
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove("./" + appName + "/go.mod")

	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		exitGracefully(err)
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), "./"+appName+"/go.mod")
	if err != nil {
		exitGracefully(err)
	}

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir("./" + appName)
	updateSource()

	// run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy...")

	cmd := exec.Command("go", "get", "gitlab.com/socle-framework/socle")
	err = cmd.Start()
	if err != nil {
		exitGracefully(err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		exitGracefully(err)
	}

	color.Green("Done building " + appURL)
	color.Green("Go build something awesome")
}

func doSkeleton(skeletonName, appName string) error {

	skeletonRepo := "https://github.com/socle-framework/arch.git"

	color.Green("\tCloning skeleton '%s' into '%s'...", skeletonName, appName)

	// git clone --filter=blob:none --sparse <repo> <dir>
	cmd := exec.Command("git", "clone", "--filter=blob:none", "--sparse", skeletonRepo, appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// git sparse-checkout set web-api
	cmd = exec.Command("git", "-C", appName, "sparse-checkout", "set", skeletonName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set sparse-checkout folder: %w", err)
	}

	srcPath := fmt.Sprintf("%s/%s", appName, skeletonName)
	if err := moveContents(srcPath, appName); err != nil {
		return fmt.Errorf("failed to move files: %w", err)
	}

	if err := os.RemoveAll(srcPath); err != nil {
		return fmt.Errorf("failed to clean temp skeleton folder: %w", err)
	}

	return nil
}

func moveContents(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := srcDir + "/" + entry.Name()
		destPath := destDir + "/" + entry.Name()

		err = os.Rename(srcPath, destPath)
		if err != nil {
			// fallback to copy + delete if rename fails (e.g., across FS boundaries)
			err = copyFileOrDir(srcPath, destPath)
			if err != nil {
				return err
			}
			_ = os.RemoveAll(srcPath)
		}
	}
	return nil
}

func copyFileOrDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.MkdirAll(dest, 0755)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
