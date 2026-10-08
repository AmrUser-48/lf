package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

type viMode uint8
const (
	viNormal viMode = iota
	viInsert
	viCommand
)

type builtinVi struct {
	lines []string
	row, col int
	top int
	mode viMode
	dirty bool
	status string
	filename string
	quit bool
	forceQuit bool
}

func runBuiltinVi(path string) error {
	b, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) { return fmt.Errorf("vi: %w", err) }
	v := &builtinVi{filename:path}
	if len(b) == 0 { v.lines=[]string{""} } else { v.lines=strings.Split(strings.ReplaceAll(string(b),"\r\n","\n"),"\n"); if len(v.lines)>1 && v.lines[len(v.lines)-1]=="" { v.lines=v.lines[:len(v.lines)-1] } }
	if len(v.lines)==0 { v.lines=[]string{""} }

	old, err := term.MakeRaw(int(os.Stdin.Fd())); if err != nil { return err }
	defer term.Restore(int(os.Stdin.Fd()), old)
	fmt.Fprint(os.Stdout, "\x1b[?25l\x1b[2J\x1b[H")
	defer fmt.Fprint(os.Stdout, "\x1b[?25h\x1b[0m\x1b[2J\x1b[H")

	in := bufio.NewReader(os.Stdin)
	for !v.quit {
		v.draw()
		key, err := readTerminalKey(in)
		if err != nil { return err }
		v.handle(key)
	}
	if v.dirty && !v.forceQuit { return nil }
	return nil
}

func readTerminalKey(r *bufio.Reader) (string,error) {
	b,err:=r.ReadByte(); if err!=nil{return "",err}
	if b==0x1b {
		b2,err:=r.ReadByte(); if err!=nil{return "esc",nil}
		if b2!='[' { if b2=='j'||b2=='k'||b2=='f'||b2=='b' { return "alt-"+string(b2),nil }; return "esc",nil }
		var seq []byte
		seq=append(seq,b2)
		for len(seq)<8 {
			c,e:=r.ReadByte(); if e!=nil{return "esc",nil}; seq=append(seq,c)
			if (c>='@'&&c<='~') { break }
		}
		s:=string(seq)
		switch s { case "[21~": return "f10",nil; case "[A":return "up",nil; case "[B":return "down",nil; case "[C":return "right",nil; case "[D":return "left",nil; case "[H":return "home",nil; case "[F":return "end",nil; case "[3~":return "delete",nil; case "[5~":return "pgup",nil; case "[6~":return "pgdn",nil }
		return "esc",nil
	}
	switch b { case 0x03,0x04:return "ctrl-c",nil; case 0x08,0x7f:return "backspace",nil; case '\n',0x0d:return "enter",nil; case 0x09:return "tab",nil }
	if b<0x20 { return "ctrl-"+string(rune(b+'a'-1)),nil }
	if b<utf8.RuneSelf { return string(b),nil }
	buf:=[]byte{b}; for !utf8.FullRune(buf){c,e:=r.ReadByte();if e!=nil{return string(buf),e};buf=append(buf,c)}
	return string(buf),nil
}

func (v *builtinVi) draw() {
	w,h,_:=term.GetSize(int(os.Stdout.Fd())); if w<20 {w=20}; if h<3 {h=3}
	if v.row<v.top {v.top=v.row}; if v.row>=v.top+h-1 {v.top=v.row-h+2}
	fmt.Fprint(os.Stdout,"\x1b[H\x1b[2J")
	for y:=0;y<h-1;y++ { i:=v.top+y; if i>=len(v.lines){break}; s:=v.lines[i]; if len([]rune(s))>w {s=string([]rune(s)[:w])}; fmt.Fprintf(os.Stdout,"~ %s\x1b[K\r\n",s) }
	mode:="NORMAL"; if v.mode==viInsert {mode="INSERT"} else if v.mode==viCommand {mode="COMMAND"}
	fmt.Fprintf(os.Stdout,"\x1b[7m%s | %s%s | %d,%d\x1b[0m\x1b[K",mode,v.filename,func()string{if v.dirty{return " [+]"};return ""}(),v.row+1,v.col+1)
	c:=v.col+2; if c>w {c=w}; fmt.Fprintf(os.Stdout,"\x1b[%d;%dH",min(v.row-v.top+1,h-1),c)
}

func min(a,b int)int{if a<b{return a};return b}

func (v *builtinVi) handle(k string) {
	if k=="f10"||k=="ctrl-c" { v.quit=true; v.forceQuit=true; return }
	if v.mode==viInsert { v.handleInsert(k); return }
	if v.mode==viCommand { v.handleCommand(k); return }
	v.handleNormal(k)
}

func (v *builtinVi) handleInsert(k string) {
	switch k {
	case "esc": v.mode=viNormal; if v.col>0{v.col--}
	case "backspace": if v.col>0 {r:=[]rune(v.lines[v.row]);r=append(r[:v.col-1],r[v.col:]...);v.lines[v.row]=string(r);v.col--;v.dirty=true}
	case "enter": r:=[]rune(v.lines[v.row]);v.lines[v.row]=string(r[:v.col]);v.lines=append(v.lines[:v.row+1],append([]string{string(r[v.col:])},v.lines[v.row+1:]...)...);v.row++;v.col=0;v.dirty=true
	case "tab": v.insertText("    ")
	default: if len(k)>0 && k[0]!='\x1b' {v.insertText(k)}
	}
}

func (v *builtinVi) insertText(s string){r:=[]rune(v.lines[v.row]);x:=[]rune(s);r=append(r[:v.col],append(x,r[v.col:]...)...);v.lines[v.row]=string(r);v.col+=len(x);v.dirty=true}

func (v *builtinVi) handleNormal(k string) {
	switch k {
	case "i":v.mode=viInsert
	case "a":if v.col<len([]rune(v.lines[v.row])){v.col++};v.mode=viInsert
	case "A":v.col=len([]rune(v.lines[v.row]));v.mode=viInsert
	case "o":v.lines=append(v.lines[:v.row+1],append([]string{""},v.lines[v.row+1:]...)...);v.row++;v.col=0;v.mode=viInsert;v.dirty=true
	case "O":v.lines=append(v.lines[:v.row],append([]string{""},v.lines[v.row:]...)...);v.col=0;v.mode=viInsert;v.dirty=true
	case "h","left":if v.col>0{v.col--}
	case "l","right":if v.col<len([]rune(v.lines[v.row]))-1{v.col++}
	case "j","down":if v.row<len(v.lines)-1{v.row++;v.col=min(v.col,len([]rune(v.lines[v.row]))-1);if v.col<0{v.col=0}}
	case "k","up":if v.row>0{v.row--;v.col=min(v.col,len([]rune(v.lines[v.row]))-1);if v.col<0{v.col=0}}
	case "0","home":v.col=0
	case "$","end":v.col=len([]rune(v.lines[v.row]));if v.col>0{v.col--}
	case "x","delete":r:=[]rune(v.lines[v.row]);if v.col<len(r){r=append(r[:v.col],r[v.col+1:]...);v.lines[v.row]=string(r);v.dirty=true}
	case "dd":v.deleteLine()
	case "D":r:=[]rune(v.lines[v.row]);if v.col<len(r){v.lines[v.row]=string(r[:v.col]);v.dirty=true}
	case "u":v.status="undo is not available in the minimal built-in vi"
	case "G":v.row=len(v.lines)-1;v.col=0
	case "Z": if v.status=="Z?" { v.save(); v.quit=true } else { v.status="Z?" }
	case ":":v.mode=viCommand;v.status=":"
	case "enter":v.row=min(v.row+1,len(v.lines)-1)
	case "pgup":v.row=max(0,v.row-10)
	case "pgdn":v.row=min(len(v.lines)-1,v.row+10)
	case "ZZ":v.save();v.quit=true
	default:
		// Handle the second character of dd and ZZ with a small pending state.
		if k=="d" {v.status="d"} else if v.status=="d" && k=="d" {v.status="";v.deleteLine()} else if k=="Z" && v.status=="Z?" {v.save();v.quit=true} else {v.status=""}
	}
}

func max(a,b int)int{if a>b{return a};return b}
func (v *builtinVi) deleteLine(){if len(v.lines)==1{v.lines[0]="";v.col=0}else{v.lines=append(v.lines[:v.row],v.lines[v.row+1:]...);if v.row>=len(v.lines){v.row=len(v.lines)-1};v.col=0};v.dirty=true}

func (v *builtinVi) handleCommand(k string) {
	if k=="esc"{v.mode=viNormal;v.status="";return}
	if k=="enter" { cmd:=v.status;v.status="";v.mode=viNormal;switch cmd{case ":q":if v.dirty{v.status="No write since last change"}else{v.quit=true};case ":q!":v.forceQuit=true;v.quit=true;case ":w":v.save();case ":wq",":x":v.save();v.quit=true};return}
	if k=="backspace" {if len(v.status)>1{v.status=v.status[:len(v.status)-1]}else{v.mode=viNormal;v.status=""};return}
	if len(k)==1 && k[0]>=0x20 {v.status+=k;return}
}

func (v *builtinVi) save(){if err:=os.WriteFile(v.filename,[]byte(strings.Join(v.lines,"\n")),0644);err!=nil{v.status="write error: "+err.Error()}else{v.dirty=false;v.status="written"}}
