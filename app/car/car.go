package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/shanghuiyang/rpi-devices/base"
	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/stianeikeland/go-rpio"
)

const (
	pinLed   = 4
	pinLight = 16
	pinIn1   = 17
	pinIn2   = 23
	pinIn3   = 27
	pinIn4   = 22
	pinENA   = 13
	pinENB   = 19
	pinBzr   = 10
	pinSG    = 18

	ipPattern = "((000.000.000.000))"
)

var (
	car         *dev.Car
	pageContext []byte
)

func main() {
	if err := rpio.Open(); err != nil {
		log.Fatalf("failed to open rpio, error: %v", err)
		return
	}
	defer rpio.Close()

	eng := dev.NewL298N(pinIn1, pinIn2, pinIn3, pinIn4, pinENA, pinENB)
	if eng == nil {
		log.Fatal("failed to new a L298N as engine, a car can't without any engine")
		return
	}

	dist := dev.NewUS100()
	if dist == nil {
		log.Printf("failed to new a HCSR04, will build a car without ultrasonic distance meter")
	}

	horn := dev.NewBuzzer(pinBzr)
	if horn == nil {
		log.Printf("failed to new a buzzer, will build a car without horns")
	}

	led := dev.NewLed(pinLed)
	if led == nil {
		log.Printf("failed to new a led, will build a car without leds")
	}

	light := dev.NewLed(pinLight)
	if light == nil {
		log.Printf("failed to new a light, will build a car without lights")
	}

	servo := dev.NewSG90(pinSG)
	if servo == nil {
		log.Printf("failed to new a sg90, will build a car without servo")
	}
	cam := dev.NewCamera()
	if cam == nil {
		log.Printf("failed to new a camera, will build a car without cameras")
	}

	car = dev.NewCar(
		dev.WithEngine(eng),
		dev.WithServo(servo),
		dev.WithDist(dist),
		dev.WithHorn(horn),
		dev.WithLed(led),
		dev.WithLight(light),
		dev.WithCamera(cam),
	)
	if car == nil {
		log.Fatal("failed to new a car")
		return
	}

	if err := loadHomePage(); err != nil {
		log.Fatalf("failed to load home page, error: %v", err)
		return
	}

	car.Start()
	log.Printf("car server started")

	base.WaitQuit(func() {
		car.Stop()
		rpio.Close()
	})

	http.HandleFunc("/", carServer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func loadHomePage() error {
	data, err := ioutil.ReadFile("car.html")
	if err != nil {
		return errors.New("internal error: failed to read car.html")
	}

	ip := base.GetIP()
	if ip == "" {
		return errors.New("internal error: failed to get ip")
	}

	rbuf := bytes.NewBuffer(data)
	wbuf := bytes.NewBuffer([]byte{})
	for {
		line, err := rbuf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		s := string(line)
		if strings.Index(s, ipPattern) >= 0 {
			s = strings.Replace(s, ipPattern, ip, 1)
		}
		wbuf.Write([]byte(s))
	}
	pageContext = wbuf.Bytes()
	return nil
}

func carServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		homePageHandler(w, r)
	case "POST":
		operationHandler(w, r)
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(pageContext)
}

func operationHandler(w http.ResponseWriter, r *http.Request) {
	op := r.FormValue("op")
	car.Do(dev.CarOp(op))
}
