package main

import (
	"bytes"
	"cmp"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
	"golang.org/x/term"
)

const previewLoadingDelay = 100 * time.Millisecond
const fileInfoDelay = 0 * time.Millisecond

type win struct {
	w, h, x, y int
}
