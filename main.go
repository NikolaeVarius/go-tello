package main

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
	// "io"
	// "io/ioutil"
	// "os"
	// // "log"
	// "os/exec"
	// "strconv"
	"sync/atomic"
	"time"
	// "gocv.io/x/gocv"
)

type pair struct {
	x float64
	y float64
}

var leftX, leftY, rightX, rightY atomic.Value

const offset = 32767.0

const (
	frameX    = 960
	frameY    = 720
	frameSize = frameX * frameY * 3
)

func main() {
	// var err error
	drone := tello.NewDriver("8888")

	joystickAdaptor := joystick.NewAdaptor()
	stick := joystick.NewDriver(joystickAdaptor, "dualshock3")
	// fmt.Println("mplayer")
	// mplayer := exec.Command("mplayer", "-fps", "25", "-cache", "8192", "-")
	// f, err := os.Create("/tmp/dat2")

	// window := gocv.NewWindow("Tello")

	// fmt.Println("ffmpg")
	// ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",
	// 	"-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	// ffmpegIn, err := ffmpeg.StdinPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// mplayerIn, err := mplayer.StdinPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("stdout")
	// ffmpegStdout, err := ffmpeg.StdoutPipe()
	// mplayerStdout, err := mplayer.StdoutPipe()

	// if err := ffmpeg.Start(); err != nil {
	// 	log.Fatal(err)
	// }
	// if err := mplayer.Start(); err != nil {
	// 	log.Fatal(err)
	// }
	// slurp, _ := ioutil.ReadAll(ffmpegStdout)
	// fmt.Printf("%s\n", slurp)

	// slurp1, _ := ioutil.ReadAll(mplayerStdout)
	// fmt.Printf("%s\n", slurp1)

	// fmt.Printf(ffmpegStdout)
	// fmt.Printf(mplayerStdout)

	work := func() {
		leftX.Store(float64(0.0))
		leftY.Store(float64(0.0))
		rightX.Store(float64(0.0))
		rightY.Store(float64(0.0))

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++")
			fmt.Println("Connected to Tello")
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++")
			// err := drone.StartVideo()
			// if err != nil {
			// 	fmt.Println(err)
			// 	return
			// }

			fmt.Println("Setting Video Encorder Rate")
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			fmt.Println("Setting Exposure")
			drone.SetExposure(0)

			gobot.Every(100*time.Millisecond, func() {
				err := drone.StartVideo()

				if err != nil {
					fmt.Println(err)
					return
				}
			})
		})

		// drone.On(tello.VideoFrameEvent, func(data interface{}) {
		// 	fmt.Println("event")
		// 	pkt := data.([]byte)
		// 	// err := ioutil.WriteFile("/tmp/dat1", pkt, 0644)

		// 	defer f.Close()

		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	n2, err := f.Write(pkt)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	fmt.Printf("wrote %d bytes\n", n2)
		// 	if _, err := mplayerIn.Write(pkt); err != nil {
		// 		fmt.Println(err)
		// 		return
		// 	}
		// 	// if _, err := ffmpegIn.Write(pkt); err != nil {
		// 	// 	fmt.Println(err)
		// 	// 	return
		// 	// }
		// })

		stick.On(joystick.TrianglePress, func(data interface{}) {
			fmt.Println("Taleoff Command")
			drone.TakeOff()
		})

		stick.On(joystick.XPress, func(data interface{}) {
			fmt.Println("Land Command")
			drone.Land()
		})

		stick.On(joystick.UpPress, func(data interface{}) {
			fmt.Println("FrontFlip")
			drone.FrontFlip()
		})

		stick.On(joystick.DownPress, func(data interface{}) {
			fmt.Println("BackFlip")
			drone.BackFlip()
		})

		stick.On(joystick.RightPress, func(data interface{}) {
			fmt.Println("RightFlip")
			drone.RightFlip()
		})

		stick.On(joystick.LeftPress, func(data interface{}) {
			fmt.Println("LeftFlip")
			drone.LeftFlip()
		})

		stick.On(joystick.LeftX, func(data interface{}) {
			val := float64(data.(int16))
			leftX.Store(val)
		})

		stick.On(joystick.LeftY, func(data interface{}) {
			val := float64(data.(int16))
			leftY.Store(val)
		})

		stick.On(joystick.RightX, func(data interface{}) {
			val := float64(data.(int16))
			rightX.Store(val)
		})

		stick.On(joystick.RightY, func(data interface{}) {
			val := float64(data.(int16))
			rightY.Store(val)
		})

		gobot.Every(10*time.Millisecond, func() {
			rightStick := getRightStick()

			switch {
			case rightStick.y < -10:
				drone.Forward(tello.ValidatePitch(rightStick.y, offset))
			case rightStick.y > 10:
				drone.Backward(tello.ValidatePitch(rightStick.y, offset))
			default:
				drone.Forward(0)
			}

			switch {
			case rightStick.x > 10:
				drone.Right(tello.ValidatePitch(rightStick.x, offset))
			case rightStick.x < -10:
				drone.Left(tello.ValidatePitch(rightStick.x, offset))
			default:
				drone.Right(0)
			}
		})

		gobot.Every(10*time.Millisecond, func() {
			leftStick := getLeftStick()
			switch {
			case leftStick.y < -10:
				drone.Up(tello.ValidatePitch(leftStick.y, offset))
			case leftStick.y > 10:
				drone.Down(tello.ValidatePitch(leftStick.y, offset))
			default:
				drone.Up(0)
			}

			switch {
			case leftStick.x > 20:
				drone.Clockwise(tello.ValidatePitch(leftStick.x, offset))
			case leftStick.x < -20:
				drone.CounterClockwise(tello.ValidatePitch(leftStick.x, offset))
			default:
				drone.Clockwise(0)
			}
		})

	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{drone, stick},
		work,
	)

	robot.Start()

	// now handle video frames from ffmpeg stream in main thread, to be macOS/Windows friendly
	// for {
	// 	buf := make([]byte, frameSize)
	// 	if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// 	// img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
	// 	// if img.Empty() {
	// 	// 	continue
	// 	// }

	// 	// window.IMShow(img)
	// 	// if window.WaitKey(1) >= 0 {
	// 	// 	break
	// 	// }
	// }

}

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}
