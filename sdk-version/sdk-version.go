package main

import (
	"fmt"
	"ztna-core/sdk-golang/ziti/sdkinfo"
)

func main() {
	_, sdkInfo := sdkinfo.GetSdkInfo()
	fmt.Printf("%s", sdkInfo.Version)
}
