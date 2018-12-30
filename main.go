package main

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
	"gocv.io/x/gocv"
	"io"
	// "io/ioutil"
	"log"
	// "os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"
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

var drone = tello.NewDriver("8888")
var window = gocv.NewWindow("Tello")
var joystickAdaptor = joystick.NewAdaptor()
var stick = joystick.NewDriver(joystickAdaptor, "dualshock3")
var ffmpeg = exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0", "-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
var ffmpegIn, _ = ffmpeg.StdinPipe()
var ffmpegOut, _ = ffmpeg.StdoutPipe()
var flightData *tello.FlightData

func init() {
	controller_listener()

	if err := ffmpeg.Start(); err != nil {
		log.Fatal(err)
	}

	drone.On(tello.FlightDataEvent, func(data interface{}) {
		flightData = data.(*tello.FlightData)
	})

}

func main() {
	fmt.Println("Starting Program")

	work := func() {
		fmt.Println("Starting Work")
		leftX.Store(float64(0.0))
		leftY.Store(float64(0.0))
		rightX.Store(float64(0.0))
		rightY.Store(float64(0.0))

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++")
			fmt.Println("Connected to Tello")
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++")

			err := drone.StartVideo()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("Setting Video Encorder Rate")
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			fmt.Println("Setting Exposure")
			drone.SetExposure(0)
			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})

		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				fmt.Println(err)
			}
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
		[]gobot.Connection{},
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{drone, stick},
		work,
	)

	robot.Start()

	for {

		buf := make([]byte, frameSize)
		if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
			fmt.Println(err)
			continue
		}

		img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
		if img.Empty() {
			continue
		}

		window.IMShow(img)
		window.WaitKey(100)
		if window.WaitKey(100) >= 0 {
			break
		}
	}
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

func controller_listener() {
	stick.On(joystick.TrianglePress, func(data interface{}) {
		fmt.Println("Takeoff Command")
		drone.TakeOff()
	})

	stick.On(joystick.XPress, func(data interface{}) {
		fmt.Println("Land Command")
		drone.Land()
	})

	stick.On(joystick.UpPress, func(data interface{}) {
		fmt.Println("Front Flip")
		drone.FrontFlip()
	})

	stick.On(joystick.DownPress, func(data interface{}) {
		fmt.Println("Back Flip")
		drone.BackFlip()
	})

	stick.On(joystick.RightPress, func(data interface{}) {
		fmt.Println("Right Flip")
		drone.RightFlip()
	})

	stick.On(joystick.LeftPress, func(data interface{}) {
		fmt.Println("Left Flip")
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
}
