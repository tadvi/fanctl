package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	sens, fan string
	timeout   int
	verbose   bool
	level     float64
)

func init() {
	flag.StringVar(&sens, "sensors", "sensors", "default command to get sensors info")
	flag.StringVar(&fan, "fan", "/sys/class/hwmon/hwmon3/pwm1", "file to write new fan speed into")

	flag.Float64Var(&level, "level", 2.6, "level")
	flag.IntVar(&timeout, "timeout", 60, "timeout in seconds")
	flag.BoolVar(&verbose, "v", false, "verbose output")
}

var re = regexp.MustCompile(`Core (\d):\s+\+(\d+)(.*)`)

func main() {
	flag.Parse()
	var last int

	for {
		time.Sleep(time.Duration(timeout) * time.Second)

		var max int

		cmd := exec.Command(sens)
		in, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
			log.Println("Defaulting to 50C")
			max = 50
		} else {
			sc := bufio.NewScanner(bytes.NewReader(in))
			for sc.Scan() {
				s := sc.Text()
				ma := re.FindStringSubmatch(s)

				//fmt.Printf("%q\n", ma)
				if len(ma) > 2 {
					//fmt.Println(ma[2])

					tp, err := strconv.Atoi(ma[2])
					if err != nil {
						log.Println(err)
						continue
					}
					if tp > max {
						max = tp
					}

				}
			}
		}

		// Check if change in tempr is significant.
		if max > last+7 || max < last-7 {
			if verbose {
				log.Println("Read max sensors temperature", max)
			}

			var t int
			t = int(float64(max)*level - 47)
			if t > 255 {
				t = 255
			}
			if t < 0 {
				t = 0
			}

			if verbose {
				log.Println("Setting fan speed to", t)
			}

			s := fmt.Sprintf("%d", t)
			ioutil.WriteFile(fan+"_enable", []byte("1"), 0644)
			ioutil.WriteFile(fan, []byte(s), 0644)

			last = max
		}

	}
}

/*
var stdio = strings.NewReader(`
asus-isa-0000
Adapter: ISA adapter
cpu_fan:       -1 RPM
temp1:       +6280.0°C

acpitz-virtual-0
Adapter: Virtual device
temp1:        +60.0°C  (crit = +104.0°C)

coretemp-isa-0000
Adapter: ISA adapter
Core 0:       +48.0°C  (high = +105.0°C, crit = +105.0°C)
Core 1:       +47.0°C  (high = +105.0°C, crit = +105.0°C)
Core 2:       +49.0°C  (high = +105.0°C, crit = +105.0°C)
Core 3:       +49.0°C  (high = +105.0°C, crit = +105.0°C)
`)
*/
