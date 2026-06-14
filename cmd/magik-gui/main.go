package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/ehsatn/MAGIK/internal/webgui"
	"github.com/ehsatn/MAGIK/pkg/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("MAGIK GUI", version.String())
		return
	}

	addr := flag.String("addr", "127.0.0.1:0", "local address for the GUI server")
	noOpen := flag.Bool("no-open", false, "do not open the browser automatically")
	flag.Parse()

	quit := make(chan struct{})
	var once sync.Once

	server := webgui.NewServer(version.Version, func() {
		once.Do(func() { close(quit) })
	})
	url, err := server.Start(*addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if !*noOpen {
		_ = openBrowser(url)
	}
	fmt.Println("MAGIK GUI:", url)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signals:
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
