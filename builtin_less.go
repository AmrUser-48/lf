package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func runBuiltinLess(path string) error {
	b,err:=os.ReadFile(path);if err!=nil{return fmt.Errorf("less: %w",err)}
	lines:=strings.Split(strings.ReplaceAll(string(b),"\r\n","\n"),"\n")
	old,err:=term.MakeRaw(int(os.Stdin.Fd()));if err!=nil{return err};defer term.Restore(int(os.Stdin.Fd()),old)
	fmt.Fprint(os.Stdout,"\x1b[?25l\x1b[2J\x1b[H");defer fmt.Fprint(os.Stdout,"\x1b[?25h\x1b[0m\x1b[2J\x1b[H")
	in:=bufio.NewReader(os.Stdin);top:=0
	for {w,h,_:=term.GetSize(int(os.Stdout.Fd()));if h<2{h=2};maxTop:=len(lines)-(h-1);if maxTop<0{maxTop=0};if top>maxTop{top=maxTop}
		fmt.Fprint(os.Stdout,"\x1b[H")
		for y:=0;y<h-1;y++{i:=top+y;if i>=len(lines){fmt.Fprint(os.Stdout,"\x1b[K\r\n");continue};s:=lines[i];if len([]rune(s))>w{s=string([]rune(s)[:w])};fmt.Fprintf(os.Stdout,"%s\x1b[K\r\n",s)}
		fmt.Fprintf(os.Stdout,"\x1b[7m %d/%d  Alt+j/k scroll  Alt+F/B page  q quit \x1b[0m\x1b[K",min(top+h-1,len(lines)),len(lines))
		k,e:=readTerminalKey(in);if e!=nil{return e}
		switch k{case "q","esc","f10":return nil;case "alt-j","down","enter":top++;case "alt-k","up":top--;case "alt-f","pgdn"," ":top+=h-1;case "alt-b","pgup","b":top-=h-1;case "home":top=0;case "end":top=maxTop}
	}
}
