package outputfile

import (
	"os"

	"github.com/bensoncb/GoScan/internal/structs/inputFile"
)

func OutputFile(ifData inputFile.InputFile) error {
	println("recieved output data")
	//Save off received data
	fil, err := os.Create("rcvd/" + ifData.Name)

	if err != nil {
		return err
	}

	defer fil.Close()

	fil.Write(ifData.Data)

	return nil
}
