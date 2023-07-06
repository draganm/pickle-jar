package picklejar_test

import (
	"fmt"
	"os"
	"testing"

	picklejar "github.com/draganm/pickle-jar"
)

func TestMain(m *testing.M) {
	// m.Run()
	err := picklejar.RunTests("features")
	if err != nil {
		fmt.Println("tests failed: ", err.Error())
		os.Exit(1)
	}
}

func TestXxx(t *testing.T) {

}
