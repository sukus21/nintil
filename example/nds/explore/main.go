package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	r := NewRunner(os.Stdin)
	addNdsFuncs(r)
	if len(os.Args) > 1 {
		r.CmdPrepend(os.Args[1:])
	}

	stdIn := bufio.NewScanner(os.Stdin)
	for {
		err := r.CmdRun()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !stdIn.Scan() {
			break
		}
		txt := stdIn.Text()
		cmd := parse(txt)
		r.CmdAppend(cmd)
	}
}

func parse(txt string) []string {
	mode := "start"
	current := ""
	out := make([]string, 0, 16)
	for _, v := range txt {
		switch mode {
		case "start":
			if v == '"' {
				mode = "quoted"
				continue
			}
			if v == ' ' {
				continue
			}
			fallthrough

		case "normal":
			if v == ' ' {
				out = append(out, current)
				current = ""
				mode = "start"
				continue
			}
			current += string(v)
			mode = "normal"

		case "quoted":
			if v == '"' {
				out = append(out, current)
				current = ""
				mode = "start"
				continue
			}
			current += string(v)
		}
	}

	if mode == "normal" && current != "" {
		out = append(out, current)
	}
	if mode == "quoted" {
		fmt.Println("syntax error, no end quote found")
		return nil
	}

	return out
}
