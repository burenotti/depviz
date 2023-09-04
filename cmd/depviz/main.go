package main

import (
	"context"
	"depviz/internal/app"
	"depviz/internal/serializers/dot"
	"depviz/internal/services/dep"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	var packageName string
	var outputFile string
	flag.StringVar(&packageName, "p", "", "package name")
	flag.StringVar(&outputFile, "o", "", "output file")
	flag.Usage = func() {
		fmt.Println("Usage of depviz")
		flag.PrintDefaults()
	}
	flag.Parse()
	if packageName == "" {
		fmt.Println("Provide package name")
		flag.Usage()
		os.Exit(1)
	}

	output, err := getOutputStream(outputFile)
	defer func(output io.WriteCloser) {
		err := output.Close()
		if err != nil {
			log.Fatalf("can't close file: %v", err)
		}
	}(output)
	if err != nil {
		log.Fatalf("can't open output file: %s", err.Error())
	}

	downloader := dep.Default()
	serializer := &dot.DotSerializer{}

	a := app.App{
		DepsProvider: downloader,
		Serializer:   serializer,
	}
	ctx := context.Background()
	if err := a.Run(ctx, packageName, output); err != nil {
		log.Fatalln(err.Error())
	}
}

func getOutputStream(outputFile string) (io.WriteCloser, error) {
	if outputFile == "" {
		return os.Stdout, nil
	}
	return os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 777)
}
