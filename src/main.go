package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	exit := 0
	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		option_s := []byte(args[0])
		option := option_s[0]
		switch option {
		case 'I':
			fmt.Println("Insert")
		case 'C':
			fmt.Println("Create tree")
		case 'F':
			fmt.Println("Find key")
		case 'T':
			fmt.Println("Display tree")
		case 'S':
			fmt.Println("Display records")
		case 'R':
			fmt.Println("Reorganize tree")
		case 'D':
			fmt.Println("Delete node")
		case 'E':
			exit = 1
		default:
			fmt.Println("Bad option")
		}
		if exit == 1 {
			break
		}
	}
}
