# GO SenseHat Model


SAMPLE FROM HERE: https://github.com/viamrobotics/rdk/blob/main/examples/customresources/demos/complexmodule/module.go


Maybe helpful: Example: https://pkg.go.dev/go.viam.com/rdk/examples/mysensor#section-readme but rather use this as a reference: https://github.com/viamrobotics/rdk/blob/main/components/sensor/sht3xd/sht3xd.go

-> SenseHat Details: https://pinout.xyz/pinout/sense_hat


simple module example: https://github.com/viamrobotics/rdk/blob/main/examples/customresources/demos/simplemodule/module.go


## Useful Go Commands

go mod init go.sensehat.senor/environment

go get go.viam.com/rdk/robot/client

go build -o bin/lps25h lps25h

go install custom/sensehat/lps25h


