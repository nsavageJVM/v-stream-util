package main
import (
	"github.com/jroimartin/gocui"
	"log"
	"github.com/magiconair/properties"
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"strings"
	"path/filepath"
	"sync"
	"os/exec"
)

const (
	command = "ffmpeg"
	codec = "h264"
	res = "scale=1280x720"
	pfmt = "-pix_fmt"
	pfmt_v = "yuv420p"
	out_cont = "mpegts"
)

type Cam struct {
	Name string
	Uri string
	Dir string
}

type App struct {
	FileList []string
	Cam1 Cam
	Cam2 Cam
	Cam3 Cam
	SelectedCam int

}


func main() {

	// make the download video directory and set up the config.properties see included example
	p := properties.MustLoadFile("${HOME}/vid/config.properties", properties.UTF8)

	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	camera1  := Cam{"cam1",p.MustGetString("cam1"), p.MustGetString("fi_1")}
	camera2  := Cam{"cam2",p.MustGetString("cam2"), p.MustGetString("fi_2") }
	camera3  := Cam{"cam3",p.MustGetString("cam3"), p.MustGetString("fi_3") }

	app := &App{Cam1: camera1, Cam2: camera2, Cam3: camera3 }

	initApp(app, g)

	app.setMainView(g)
	app.SetKeys(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}



func initApp(a *App, g *gocui.Gui) {
	g.Cursor = true
	g.InputEsc = false
	g.BgColor = gocui.ColorCyan
	g.FgColor = gocui.ColorBlack

}


func (app *App) setMainView(g *gocui.Gui ) error {
	maxX, maxY := g.Size()

	if v, err :=  g.SetView("menu", 2, 1, maxX/7, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "  V-Stream Camera Setup Menu "
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "Commands")
		fmt.Fprintln(v, "Run Cam1")
		fmt.Fprintln(v, "Run Cam2")
		fmt.Fprintln(v, "Run Cam3")
		fmt.Fprintln(v, "Quit")

	}

	if v, err :=  g.SetView("main", maxX/7 +2, 1, maxX -10, maxY -4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title ="  V_STREAM Camera Manager "

		//app.info(v)

		v.Wrap = true
		if _, err := g.SetCurrentView("menu"); err != nil {
			return err
		}

	}
	return nil
}


func (app *App) popup(g *gocui.Gui, msg string ) {
	var popup *gocui.View
	var err error
	maxX, maxY := g.Size()
	if popup, err = g.SetView("popup", maxX/3-len(msg)/2-1, maxY/3-1, maxX/3+len(msg)/2+4, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return
		}
		popup.Frame = true
		popup.Wrap = true
		popup.Title = "Info"
		popup.Clear()
		fmt.Fprint(popup, msg)
		popup.SetCursor(len(msg), 0)
		g.SetViewOnTop("popup")
		g.SetCurrentView("popup")
	}
}


func (app *App) SetKeys(g *gocui.Gui) {

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, app.quit); err != nil { }

	if err := g.SetKeybinding("menu", gocui.KeyArrowDown, gocui.ModNone, app.cursorChange); err != nil {  }

	if err := g.SetKeybinding("menu", gocui.KeyArrowUp, gocui.ModNone, app.cursorChange); err != nil {  }

	if err := g.SetKeybinding("popup", gocui.KeyEnter, gocui.ModNone, app.runexec); err != nil {  }

	if err := g.SetKeybinding("popup", gocui.KeyEsc, gocui.ModNone, app.closePopUp); err != nil {  }

}


func (app *App) closePopUp(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("popup"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("menu"); err != nil {
		return err
	}
	return nil
}


func (app *App) quit(g *gocui.Gui, v *gocui.View) error {

	return gocui.ErrQuit
}

func (app *App) cursorChange(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}

	app.updateMainView(g, v)

	return nil
}


func (app *App) updateMainView(g *gocui.Gui, v *gocui.View) {

	var l string
	var err error

	_, cy := v.Cursor()

	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	v2, err := g.SetCurrentView("main")
	if err != nil {
		log.Panicln(err)
	}
	v2.Clear()

	if (l == "Run Cam1") {
		app.SelectedCam =1
		app.popup(g, "do you want run cam1:\nEsc to return to app\nEnter to run")

	} else if ( l == "Run Cam2") {
		app.SelectedCam =2
		app.popup(g, "do you want run cam2:\nEsc to return to app\nEnter to run")

	} else if (l == "Run Cam3") {
		app.SelectedCam =3
		app.popup(g, "do you want run cam3:\nEsc to return to app\nEnter to run")

	}   else if (l == "Quit") {
		fmt.Fprintln(v2, "Quitting ......")
		termbox.Close()
		os.Exit(0)
	}

}

func (app *App) runexec(g *gocui.Gui, v *gocui.View) error {

	if err := g.DeleteView("popup"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("menu"); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Fprintln(v, "Setting up camera %v please be patient ......", app.SelectedCam )
	wg.Done()

	//https://github.com/golang/go/wiki/Switch
	switch x :=  app.SelectedCam;  true {

	case x == 1:

		app.deleteFile(app.Cam1.Dir )
		app.runCamera(app.Cam1.Uri, app.Cam1.Dir)

	case x == 2:

		app.deleteFile(app.Cam2.Dir)
		app.runCamera(app.Cam2.Uri, app.Cam2.Dir)

	case x == 3:

		app.deleteFile(app.Cam3.Dir)
		app.runCamera(app.Cam3.Uri, app.Cam3.Dir)

	}

	return nil
}



func (app *App) deleteFile(path string ) error {

	err :=  os.Remove(path)
	if err != nil {

	}

	return nil
}


func (app *App) runCamera(cam_url string, cam_file string )  {


	go func( ) {
		cmd := exec.Command(command, "-i", cam_url, "-c:v", codec, "-vf", res, "-an", pfmt, pfmt_v, "-f", out_cont, cam_file)
		//	cmd := exec.Command(command, "-i", cam_url, "-c:v", codec, "-vf", res, "-an", pfmt, pfmt_v,  cam_file)
		if err := cmd.Start(); err != nil {
			log.Fatal(err)

		}

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)

		}

	}()

}




func (app *App) findFiles(path string, f os.FileInfo, err error) error {

	if strings.Contains(path, "Trash") {
		return nil
	}

	if filepath.Ext(path) == ".mp4" {
		app.FileList = append(app.FileList, path)

	}

	return nil
}