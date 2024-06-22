package deej

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/micmonay/keybd_event"
)

func PressButton(cmds []string) {
	for _, cmd := range cmds {
		cmdArray := strings.Split(cmd, ".")
		switch cmdArray[0] {
		case "keyboard":
			pressKey(cmdArray[1])
		default:
			openFile(cmd)
		}
	}
}

func openFile(filePath string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	default: // Linux, BSD, etc.
		args := strings.Fields(filePath)
		cmd = exec.Command(args[0], args[1:]...)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("failed to open file: %v", err)
	}
}

func pressKey(s string) {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	fmt.Println(keyMap[s])
	kb.SetKeys(keyMap[s])
	if err := kb.Launching(); err != nil {
		panic(err)
	}
}

var keyMap = map[string]int{
	"lxnextsong":     163,
	"lxplaypause":    164,
	"lxprevioussong": 165,
	"lxmute":         113,
	"lxvoldown":      114,
	"lxvolup":        115,
	"lxpause":        119,
	//Windows Keycodes below
	"shift":               0x10 + 0xFFF,
	"ctrl":                0x11 + 0xFFF,
	"alt":                 0x12 + 0xFFF,
	"lshift":              0xA0 + 0xFFF,
	"rshift":              0xA1 + 0xFFF,
	"lcontrol":            0xA2 + 0xFFF,
	"rcontrol":            0xA3 + 0xFFF,
	"lwin":                0x5B + 0xFFF,
	"rwin":                0x5C + 0xFFF,
	"key_up":              0x0002,
	"scan_code":           0x0008,
	"sp1":                 41,
	"sp2":                 12,
	"sp3":                 13,
	"sp4":                 26,
	"sp5":                 27,
	"sp6":                 39,
	"sp7":                 40,
	"sp8":                 43,
	"sp9":                 51,
	"sp10":                52,
	"sp11":                53,
	"sp12":                86,
	"esc":                 1,
	"1":                   2,
	"2":                   3,
	"3":                   4,
	"4":                   5,
	"5":                   6,
	"6":                   7,
	"7":                   8,
	"8":                   9,
	"9":                   10,
	"0":                   11,
	"q":                   16,
	"w":                   17,
	"e":                   18,
	"r":                   19,
	"t":                   20,
	"y":                   21,
	"u":                   22,
	"i":                   23,
	"o":                   24,
	"p":                   25,
	"a":                   30,
	"s":                   31,
	"d":                   32,
	"f":                   33,
	"g":                   34,
	"h":                   35,
	"j":                   36,
	"k":                   37,
	"l":                   38,
	"z":                   44,
	"x":                   45,
	"c":                   46,
	"v":                   47,
	"b":                   48,
	"n":                   49,
	"m":                   50,
	"f1":                  59,
	"f2":                  60,
	"f3":                  61,
	"f4":                  62,
	"f5":                  63,
	"f6":                  64,
	"f7":                  65,
	"f8":                  66,
	"f9":                  67,
	"f10":                 68,
	"f11":                 87,
	"f12":                 88,
	"f13":                 0x7C + 0xFFF,
	"f14":                 0x7D + 0xFFF,
	"f15":                 0x7E + 0xFFF,
	"f16":                 0x7F + 0xFFF,
	"f17":                 0x80 + 0xFFF,
	"f18":                 0x81 + 0xFFF,
	"f19":                 0x82 + 0xFFF,
	"f20":                 0x83 + 0xFFF,
	"f21":                 0x84 + 0xFFF,
	"f22":                 0x85 + 0xFFF,
	"f23":                 0x86 + 0xFFF,
	"f24":                 0x87 + 0xFFF,
	"numlock":             69,
	"scrolllock":          70,
	"reserved":            0,
	"minus":               12,
	"equal":               13,
	"backspace":           14,
	"tab":                 15,
	"leftbrace":           26,
	"rightbrace":          27,
	"enter":               28,
	"semicolon":           39,
	"apostrophe":          40,
	"grave":               41,
	"backslash":           43,
	"comma":               51,
	"dot":                 52,
	"slash":               53,
	"kp_asterisk":         55,
	"space":               57,
	"capslock":            58,
	"kp0":                 82,
	"kp1":                 79,
	"kp2":                 80,
	"kp3":                 81,
	"kp4":                 75,
	"kp5":                 76,
	"kp6":                 77,
	"kp7":                 71,
	"kp8":                 72,
	"kp9":                 73,
	"kp_minus":            74,
	"kp_plus":             78,
	"kp_dot":              83,
	"lbutton":             0x01 + 0xFFF,
	"rbutton":             0x02 + 0xFFF,
	"cancel":              0x03 + 0xFFF,
	"mbutton":             0x04 + 0xFFF,
	"xbutton1":            0x05 + 0xFFF,
	"xbutton2":            0x06 + 0xFFF,
	"back":                0x08 + 0xFFF,
	"clear":               0x0C + 0xFFF,
	"pause":               0x13 + 0xFFF,
	"capital":             0x14 + 0xFFF,
	"kana":                0x15 + 0xFFF,
	"hanguel":             0x15 + 0xFFF,
	"hangul":              0x15 + 0xFFF,
	"junja":               0x17 + 0xFFF,
	"final":               0x18 + 0xFFF,
	"hanja":               0x19 + 0xFFF,
	"kanji":               0x19 + 0xFFF,
	"convert":             0x1C + 0xFFF,
	"nonconvert":          0x1D + 0xFFF,
	"accept":              0x1E + 0xFFF,
	"modechange":          0x1F + 0xFFF,
	"pageup":              0x21 + 0xFFF,
	"pagedown":            0x22 + 0xFFF,
	"end":                 0x23 + 0xFFF,
	"home":                0x24 + 0xFFF,
	"left":                0x25 + 0xFFF,
	"up":                  0x26 + 0xFFF,
	"right":               0x27 + 0xFFF,
	"down":                0x28 + 0xFFF,
	"select":              0x29 + 0xFFF,
	"print":               0x2A + 0xFFF,
	"execute":             0x2B + 0xFFF,
	"snapshot":            0x2C + 0xFFF,
	"insert":              0x2D + 0xFFF,
	"delete":              0x2E + 0xFFF,
	"help":                0x2F + 0xFFF,
	"scroll":              0x91 + 0xFFF,
	"lmenu":               0xA4 + 0xFFF,
	"rmenu":               0xA5 + 0xFFF,
	"browser_back":        0xA6 + 0xFFF,
	"browser_forward":     0xA7 + 0xFFF,
	"browser_refresh":     0xA8 + 0xFFF,
	"browser_stop":        0xA9 + 0xFFF,
	"browser_search":      0xAA + 0xFFF,
	"browser_favorites":   0xAB + 0xFFF,
	"browser_home":        0xAC + 0xFFF,
	"volume_mute":         0xAD + 0xFFF,
	"volume_down":         0xAE + 0xFFF,
	"volume_up":           0xAF + 0xFFF,
	"media_next_track":    0xB0 + 0xFFF,
	"media_prev_track":    0xB1 + 0xFFF,
	"media_stop":          0xB2 + 0xFFF,
	"media_play_pause":    0xB3 + 0xFFF,
	"launch_mail":         0xB4 + 0xFFF,
	"launch_media_select": 0xB5 + 0xFFF,
	"launch_app1":         0xB6 + 0xFFF,
	"launch_app2":         0xB7 + 0xFFF,
	"oem_1":               0xBA + 0xFFF,
	"oem_plus":            0xBB + 0xFFF,
	"oem_comma":           0xBC + 0xFFF,
	"oem_minus":           0xBD + 0xFFF,
	"oem_period":          0xBE + 0xFFF,
	"oem_2":               0xBF + 0xFFF,
	"oem_3":               0xC0 + 0xFFF,
	"oem_4":               0xDB + 0xFFF,
	"oem_5":               0xDC + 0xFFF,
	"oem_6":               0xDD + 0xFFF,
	"oem_7":               0xDE + 0xFFF,
	"oem_8":               0xDF + 0xFFF,
	"oem_102":             0xE2 + 0xFFF,
	"processkey":          0xE5 + 0xFFF,
	"packet":              0xE7 + 0xFFF,
	"attn":                0xF6 + 0xFFF,
	"crsel":               0xF7 + 0xFFF,
	"exsel":               0xF8 + 0xFFF,
	"ereof":               0xF9 + 0xFFF,
	"play":                0xFA + 0xFFF,
	"zoom":                0xFB + 0xFFF,
	"noname":              0xFC + 0xFFF,
	"pa1":                 0xFD + 0xFFF,
	"oem_clear":           0xFE + 0xFFF,
}
