package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/tajtiattila/xinput"
	"github.com/vence722/inputhook"
	"github.com/xackery/w32"
	"math"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

var savestateKey = false
var loadsateKey = false
var handle w32.HANDLE

func GetProcessName(id uint32) string {
	snapshot := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPMODULE, id)
	if snapshot == w32.ERROR_INVALID_HANDLE {
		return "<UNKNOWN>"
	}
	defer w32.CloseHandle(snapshot)

	var me w32.MODULEENTRY32
	me.Size = uint32(unsafe.Sizeof(me))
	if w32.Module32First(snapshot, &me) {
		return w32.UTF16PtrToString(&me.SzModule[0])
	}

	return "<UNKNOWN>"
}

func ListProcesses() []uint32 {
	sz := uint32(1000)
	procs := make([]uint32, sz)
	var bytesReturned uint32
	if w32.EnumProcesses(procs, sz, &bytesReturned) {
		return procs[:int(bytesReturned)/4]
	}
	return []uint32{}
}

func FindProcessByName(name string) (uint32, error) {
	for _, pid := range ListProcesses() {
		if GetProcessName(pid) == name {
			return pid, nil
		}
	}
	return 0, fmt.Errorf("unknown process")
}
func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
func Float32bytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func findName(location uint32, handle w32.HANDLE) string {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000044")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-1")
		fmt.Println(err)
		return ""
	} else {
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}

	name, err := w32.ReadProcessMemory(handle, nameLoc, uint(32))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-1")
		fmt.Println(err)
		return ""
	} else {
		return string(name)
		//fmt.Println(strconv.Itoa(data))

	}
}

func getXValue(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 20
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(8))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		fmt.Print("x Loc: ")
		fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}

func getYValue(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 24
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(8))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		fmt.Print("Y Loc: ")
		fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func getZValue(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 28
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(8))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		fmt.Print("Z Loc: ")
		fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func getPhysics(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {

	}
	var dataOffset uint32
	dataOffset = 0xC0
	physics, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(0x84))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return physics
		//fmt.Println(strconv.Itoa(data))

	}
}
func setPhysics(location uint32, handle w32.HANDLE, physics []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0xC0
	err = w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, physics, uint(0x84))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		return
		//fmt.Println(strconv.Itoa(data))

	}
}

func getAnimFrames(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x174
	animFrames, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return animFrames
		//fmt.Println(strconv.Itoa(data))

	}
}

func setAnimFrames(location uint32, handle w32.HANDLE, animFrames []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x174
	err = w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, animFrames, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		return
		//fmt.Println(strconv.Itoa(data))

	}
}

func getCurrAnim(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x176
	currAnim, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(6))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return currAnim
		//fmt.Println(strconv.Itoa(data))

	}
}

func setCurrAnim(location uint32, handle w32.HANDLE, currAnim []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x176
	err = w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, currAnim, uint(6))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		return
		//fmt.Println(strconv.Itoa(data))

	}
}

func getMomentum(handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	pointerToMomentum, pointerToMomentumErr := w32.HexToUint32("01A51A9C")
	if pointerToMomentumErr != nil {
		fmt.Println(pointerToMomentumErr)
	}

	momentumLocation, err := w32.ReadProcessMemoryAsUint32(handle, pointerToMomentum)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x44
	momentum, err := w32.ReadProcessMemory(handle, momentumLocation+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(momentum)
		//fmt.Println(strconv.Itoa(data))

	}
}
func setMomentum(handle w32.HANDLE, momentum float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	pointerToMomentum, pointerToMomentumErr := w32.HexToUint32("01A51A9C")
	if pointerToMomentumErr != nil {
		fmt.Println(pointerToMomentumErr)
	}

	momentumLocation, err := w32.ReadProcessMemoryAsUint32(handle, pointerToMomentum)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x44
	err = w32.WriteProcessMemory(handle, momentumLocation+dataOffset, Float32bytes(momentum), uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}

func getHover(handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	pointerToHover, pointerToHoverErr := w32.HexToUint32("01A51A9C")
	if pointerToHoverErr != nil {
		fmt.Println(pointerToHoverErr)
	}

	hoverLocation, err := w32.ReadProcessMemoryAsUint32(handle, pointerToHover)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x12
	hover, err := w32.ReadProcessMemory(handle, hoverLocation+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return hover
		//fmt.Println(strconv.Itoa(data))

	}
}

func setHover(handle w32.HANDLE, hover []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	pointerToHover, pointerToHoverErr := w32.HexToUint32("01A51A9C")
	if pointerToHoverErr != nil {
		fmt.Println(pointerToHoverErr)
	}

	hoverLocation, err := w32.ReadProcessMemoryAsUint32(handle, pointerToHover)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)

	} else {

	}
	var dataOffset uint32
	dataOffset = 0x12
	err = w32.WriteProcessMemory(handle, hoverLocation+dataOffset, hover, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func set99Lives(handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	err := w32.WriteProcessMemory(handle, 0x174B024, []byte{99}, uint(2))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func getCamera1X(handle w32.HANDLE) float32 {

	var dataOffset uint32
	dataOffset = 0x14
	cam1X, err := w32.ReadProcessMemory(handle, 31260864+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(cam1X)
		//fmt.Println(strconv.Itoa(data))

	}
}
func getCamera1Y(handle w32.HANDLE) float32 {

	var dataOffset uint32
	dataOffset = 0x18
	cam1Y, err := w32.ReadProcessMemory(handle, 31260864+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(cam1Y)
		//fmt.Println(strconv.Itoa(data))

	}
}
func getCamera1Z(handle w32.HANDLE) float32 {

	var dataOffset uint32
	dataOffset = 0x1C
	cam1Z, err := w32.ReadProcessMemory(handle, 31260864+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(cam1Z)
		//fmt.Println(strconv.Itoa(data))

	}
}
func setCamera1X(handle w32.HANDLE, camX float32) {

	var dataOffset uint32
	dataOffset = 0x14
	err := w32.WriteProcessMemory(handle, 31260864+dataOffset, Float32bytes(camX), uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func setCamera1Y(handle w32.HANDLE, camY float32) {

	var dataOffset uint32
	dataOffset = 0x18
	err := w32.WriteProcessMemory(handle, 31260864+dataOffset, Float32bytes(camY), uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func setCamera1Z(handle w32.HANDLE, camZ float32) {

	var dataOffset uint32
	dataOffset = 0x1C
	err := w32.WriteProcessMemory(handle, 31260864+dataOffset, Float32bytes(camZ), uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func getTime(handle w32.HANDLE) []byte {
	time, err := w32.ReadProcessMemory(handle, 0x0174AFDB, uint(3))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return time
		//fmt.Println(strconv.Itoa(data))

	}
}
func setTime(handle w32.HANDLE, time []byte) {
	err := w32.WriteProcessMemory(handle, 0x0174AFDB, time, uint(3))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)

	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}

func getGravity(handle w32.HANDLE) []byte {
	grav, err := w32.ReadProcessMemory(handle, 0x01DE94A0, uint(12))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return grav
		//fmt.Println(strconv.Itoa(data))

	}
}

func setGravity(handle w32.HANDLE, grav []byte) {
	err := w32.WriteProcessMemory(handle, 0x01DE94A0, grav, uint(12))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)

	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}

func getCamera2(handle w32.HANDLE) []byte {

	var dataOffset uint32
	dataOffset = 0x4000
	cam2, err := w32.ReadProcessMemory(handle, 31260864-dataOffset, uint(40000))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return cam2
		//fmt.Println(strconv.Itoa(data))

	}
}
func setCamera2(handle w32.HANDLE, cam []byte) {

	var dataOffset uint32
	dataOffset = 0x4000
	err := w32.WriteProcessMemory(handle, 31260864-dataOffset, cam, uint(40000))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func getHangTime(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x12
	hangtime, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return hangtime
		//fmt.Println(strconv.Itoa(data))

	}
}
func setHangTime(location uint32, handle w32.HANDLE, hangtime []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x12
	err = w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, hangtime, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func getHorSpeed(location uint32, handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x64
	horSpeed, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(horSpeed)
		//fmt.Println(strconv.Itoa(data))

	}
}
func getVertSpeed(location uint32, handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x68
	vertSpeed, err := w32.ReadProcessMemory(handle, ObjectData2Location+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(vertSpeed)
		//fmt.Println(strconv.Itoa(data))

	}
}
func setHorSpeed(location uint32, handle w32.HANDLE, speed float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x64
	var speedbytes []byte
	speedbytes = Float32bytes(speed)

	err2 := w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, speedbytes, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func setVertSpeed(location uint32, handle w32.HANDLE, speed float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData2Offset uint32
	var ObjectData2OffsetError error
	var ObjectData2OffsetPointer uint32
	var ObjectData2Location uint32
	ObjectData2Offset, ObjectData2OffsetError = w32.HexToUint32("00000040")
	if ObjectData2OffsetError != nil {
		fmt.Println(ObjectData2OffsetError)
	}
	ObjectData2OffsetPointer = location + ObjectData2Offset
	ObjectData2Location, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData2OffsetPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {

	}
	var dataOffset uint32
	dataOffset = 0x68
	var speedbytes []byte
	speedbytes = Float32bytes(speed)

	err2 := w32.WriteProcessMemory(handle, ObjectData2Location+dataOffset, speedbytes, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func getStatus(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x4
	status, err := w32.ReadProcessMemory(handle, ObjectData1Loc+dataOffset, uint(2))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return status
		//fmt.Println(strconv.Itoa(data))

	}
}
func setStatus(location uint32, handle w32.HANDLE, status []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x4
	err = w32.WriteProcessMemory(handle, ObjectData1Loc+dataOffset, status, uint(2))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}
func getAction(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x0
	action, err := w32.ReadProcessMemory(handle, ObjectData1Loc+dataOffset, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return action
		//fmt.Println(strconv.Itoa(data))

	}
}

func setAction(location uint32, handle w32.HANDLE, action []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x0
	err = w32.WriteProcessMemory(handle, ObjectData1Loc+dataOffset, action, uint(1))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {

		//fmt.Println(strconv.Itoa(data))

	}
}

func getXRot(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x8
	rot, err := w32.ReadProcessMemory(handle, ObjectData1Loc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return rot
		//fmt.Println(strconv.Itoa(data))

	}
}
func getYRot(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0xC
	rot, err := w32.ReadProcessMemory(handle, ObjectData1Loc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return rot
		//fmt.Println(strconv.Itoa(data))

	}
}
func getZRot(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x10
	rot, err := w32.ReadProcessMemory(handle, ObjectData1Loc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {
		return rot
		//fmt.Println(strconv.Itoa(data))

	}
}

func setXRot(location uint32, handle w32.HANDLE, rot []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x8
	err = w32.WriteProcessMemory(handle, ObjectData1Loc+dataOffset, rot, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	}
}
func setYRot(location uint32, handle w32.HANDLE, rot []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0xC
	err = w32.WriteProcessMemory(handle, ObjectData1Loc+dataOffset, rot, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	}
}
func setZRot(location uint32, handle w32.HANDLE, rot []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var ObjectData1Offset uint32
	var ObjectData1OffsetError error
	var ObjectData1Pointer uint32
	var ObjectData1Loc uint32
	ObjectData1Offset, ObjectData1OffsetError = w32.HexToUint32("00000034")
	if ObjectData1OffsetError != nil {
		fmt.Println(ObjectData1OffsetError)
	}
	ObjectData1Pointer = location + ObjectData1Offset
	ObjectData1Loc, err := w32.ReadProcessMemoryAsUint32(handle, ObjectData1Pointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x10
	err = w32.WriteProcessMemory(handle, ObjectData1Loc+dataOffset, rot, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	}
}
func getX(location uint32, handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 20
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return Float32frombytes(name)
		//fmt.Println(strconv.Itoa(data))

	}
}

func getY(location uint32, handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 24
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {
		return (Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func getZ(location uint32, handle w32.HANDLE) float32 {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return 0
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 28
	name, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(4))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return 0
	} else {

		return (Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func getScale(location uint32, handle w32.HANDLE) []byte {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return []byte{0}
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x20
	scale, err := w32.ReadProcessMemory(handle, nameLoc+dataOffset, uint(12))
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return []byte{0}
	} else {

		return scale
		//fmt.Println(strconv.Itoa(data))

	}
}
func setScale(location uint32, handle w32.HANDLE, scale []byte) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x20

	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, scale, uint(12))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func flatX(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x20

	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, []byte{0}, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func flatY(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x24

	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, []byte{0}, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func flatZ(location uint32, handle w32.HANDLE) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 0x28

	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, []byte{0}, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func setXValue(location uint32, handle w32.HANDLE, xLoc float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 20
	var locBytes []byte
	locBytes = Float32bytes(xLoc)
	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, locBytes, uint(8))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func setYValue(location uint32, handle w32.HANDLE, yLoc float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 24
	var locBytes []byte
	locBytes = Float32bytes(yLoc)
	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, locBytes, uint(8))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func setZValue(location uint32, handle w32.HANDLE, zLoc float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 28
	var locBytes []byte
	locBytes = Float32bytes(zLoc)
	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset, locBytes, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}
func twoDSonic(location uint32, handle w32.HANDLE, zLoc float32) {

	//fmt.Println("Object Location: " + fmt.Sprint(location))

	var nameOffset uint32
	var nameOffsetError error
	var nameLocPointer uint32
	var nameLoc uint32
	nameOffset, nameOffsetError = w32.HexToUint32("00000034")
	if nameOffsetError != nil {
		fmt.Println(nameOffsetError)
	}
	nameLocPointer = location + nameOffset
	nameLoc, err := w32.ReadProcessMemoryAsUint32(handle, nameLocPointer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-100: Error Reading Object 2 Pointer")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Found Pointer: ")
		//fmt.Println(nameLoc)
		//fmt.Println(reflect.TypeOf(nameLoc))
		//fmt.Println(string(nameLoc))
		//fmt.Println(strconv.Itoa(data))

	}
	var dataOffset uint32
	dataOffset = 28
	var locBytes []byte
	locBytes = Float32bytes(zLoc)
	err2 := w32.WriteProcessMemory(handle, nameLoc+dataOffset+4, locBytes, uint(4))
	if err2 != nil {
		//Error Reading Memory: -1
		fmt.Println("-200: Error Reading Object 2 Data")
		fmt.Println(err)
		return
	} else {
		//fmt.Print("Location Set")
		//fmt.Println(Float32frombytes(name))
		//fmt.Println(strconv.Itoa(data))

	}
}

func checkIfNew(item string, seenList [8]string) bool {
	for i := 0; i < len(seenList); i++ {
		if strings.Contains(item, seenList[i]) {
			return false
		}
	}
	return true
}

func hookCallback(keyEvent int, keyCode int) {
	if keyEvent == 257 {
		if keyCode == 37 {
			savestateKey = true
		}
		if keyCode == 39 {
			loadsateKey = true
		}
	}
	//fmt.Println("keyEvent:", keyEvent)
	//fmt.Println("keyCode:", keyCode)
}
func keyboardThread() {
	inputhook.HookKeyboard(hookCallback)
	ch := make(chan bool)
	<-ch
}
func main() {

	//var state *syscall.Proc

	args := os.Args[1:]
	var command string
	if len(args) != 0 {
		command = args[0]
	}
	oldObjects := [8]string{"Extra_Exec", "epTaskExec", "LandManager", "ParticleCoreTask", "BgExec", "FogtaskMan", "MinimalCounter", "Ring"}
	var sonicPID uint32
	var err error
	sonicPID, err = FindProcessByName("sonic2app.exe")
	handle, err = w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, sonicPID)

	if err != nil {
		sonicPID, err = FindProcessByName("sonic2app")
		handle, err = w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, sonicPID)
		//Error Opening Process: -2
		fmt.Println("-2")
		w32.MessageBox(0, "Sanic not found, RIP IN PEP", "OPEN THE GAME FIRST DOOD", 0)
		return

	}

	switch command {
	case "-stage":
		var stageMemoryAddress uint32
		var stageError error

		stageMemoryAddress, stageError = w32.HexToUint32("01934B70")

		if stageError != nil {
			//Error getting Stage Stage Uint32: -3
			fmt.Println("-3")
			return
		}

		data, err := w32.ReadProcessMemoryAsUint32(handle, stageMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		} else {
			fmt.Println(data)
		}
	case "-settime":
		var timeMemoryAddress uint32
		var timeError error

		timeMemoryAddress, timeError = w32.HexToUint32("0174AFDC")

		if timeError != nil {
			//Error getting time Uint32: -3
			fmt.Println("-30")
			return
		}

		var time uint32 = 8
		timeMemoryAddress, timeError = w32.HexToUint32("0174AFDC")
		err = w32.WriteProcessMemoryAsUint32(handle, timeMemoryAddress, time)
		if err != nil {
			fmt.Println("Err: " + err.Error())
		} else {
			fmt.Println("Success")
		}
	case "-char":
		var testMemoryAddress uint32
		var offset uint32
		var offsetError error
		var testError error

		fmt.Println("Master Pointer p1 HEX: 01DEA6E0")
		testMemoryAddress, testError = w32.HexToUint32("01DEA6E0")
		if testError != nil {
			//Error getting Stage Stage Uint32: -3
			fmt.Println("-3")
			return
		}
		fmt.Println("Master Pointer p1 unit32: " + fmt.Sprint(testMemoryAddress))
		pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		fmt.Println("Master Pointer points to: " + fmt.Sprint(pointsTo))
		fmt.Println("Type of Object Pointed to:")
		fmt.Println(reflect.TypeOf(pointsTo))

		fmt.Println("Offset: 00000044")
		offset, offsetError = w32.HexToUint32("00000044")
		if offsetError != nil {
			//Error Getting Offset: -10
			fmt.Println("-10")
			return
		}
		fmt.Println("Offset as uint32: " + fmt.Sprint(offset))

		var objectPointer uint32

		objectPointer = pointsTo + offset
		fmt.Println("Object Location Pointer + offset: " + fmt.Sprint(objectPointer))
		objectLoc, err := w32.ReadProcessMemoryAsUint32(handle, objectPointer)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		}
		fmt.Println("Object Location: " + fmt.Sprint(objectLoc))

		fmt.Println("Object Data Offset Hex: 00000000")
		var objectDataOffset uint32
		var objectDataOffsetError error
		objectDataOffset, objectDataOffsetError = w32.HexToUint32("00000000")
		if objectDataOffsetError != nil {
			//Error setting Object Offset: -11
			fmt.Println("-11")
			return
		}
		fmt.Println("Object Data Offset uint32: " + fmt.Sprint(objectDataOffset))
		var objectDataLocation uint32
		objectDataLocation = objectLoc + objectDataOffset
		fmt.Println("Object Data Location uint32: " + fmt.Sprint(objectDataLocation))

		data, err := w32.ReadProcessMemory(handle, objectDataLocation, uint(16))
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(data))
			//fmt.Println(strconv.Itoa(data))

		}
		/*
			var objectOffset uint32
			var objectOffsetError error
			objectOffset, objectOffsetError = w32.HexToUint32("00000001")

			var objectSpot uint32
			objectSpot = pointsToUint + objectOffset - objectOffset
			if objectOffsetError != nil {
				//Error setting Object Offset: -11
				fmt.Println("-11")
				return
			}
			fmt.Println(objectSpot)

			data, err := w32.ReadProcessMemoryAsUint32(handle, objectSpot)
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				fmt.Println(err)
				return
			} else {
				fmt.Println(data)
			}
		*/

	case "-charobj":

		var testMemoryAddress uint32
		var offset uint32
		var offsetError error
		var testError error

		fmt.Println("Master Pointer p1 HEX: 01DEA6E0")
		testMemoryAddress, testError = w32.HexToUint32("01DEA6E0")
		if testError != nil {
			//Error getting Stage Stage Uint32: -3
			fmt.Println("-3")
			return
		}
		fmt.Println("Master Pointer p1 unit32: " + fmt.Sprint(testMemoryAddress))
		pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		fmt.Println("Master Pointer points to: " + fmt.Sprint(pointsTo))
		fmt.Println("Type of Object Pointed to:")
		fmt.Println(reflect.TypeOf(pointsTo))

		fmt.Println("Offset: 00000040")
		offset, offsetError = w32.HexToUint32("00000040")
		if offsetError != nil {
			//Error Getting Offset: -10
			fmt.Println("-10")
			return
		}
		fmt.Println("Offset as uint32: " + fmt.Sprint(offset))

		var objectPointer uint32

		objectPointer = pointsTo + offset
		fmt.Println("Object Location Pointer + offset: " + fmt.Sprint(objectPointer))
		objectLoc, err := w32.ReadProcessMemoryAsUint32(handle, objectPointer)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		}
		fmt.Println("Object Location: " + fmt.Sprint(objectLoc))

		fmt.Println("Object Data Offset Hex: 00000003")
		var objectDataOffset uint32
		var objectDataOffsetError error
		objectDataOffset, objectDataOffsetError = w32.HexToUint32("00000003")
		if objectDataOffsetError != nil {
			//Error setting Object Offset: -11
			fmt.Println("-11")
			return
		}
		fmt.Println("Object Data Offset uint32: " + fmt.Sprint(objectDataOffset))
		var objectDataLocation uint32
		objectDataLocation = objectLoc + objectDataOffset
		fmt.Println("Object Data Location uint32: " + fmt.Sprint(objectDataLocation))

		data, err := w32.ReadProcessMemory(handle, objectDataLocation, uint(1))
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		} else {
			fmt.Println(data)
			//fmt.Println(string(data))
			//fmt.Println(strconv.Itoa(data))

		}
		var character uint32 = 2
		err = w32.WriteProcessMemoryAsUint32(handle, objectDataLocation, character)
		if err != nil {
			fmt.Println("Err: " + err.Error())
		} else {
			fmt.Println("Success")
		}

	case "-getX":

		var testMemoryAddress uint32
		var offset uint32
		var offsetError error
		var testError error

		fmt.Println("Master Pointer p1 HEX: 01DEA6E0")
		testMemoryAddress, testError = w32.HexToUint32("01DEA6E0")
		if testError != nil {
			//Error getting Stage Stage Uint32: -3
			fmt.Println("-3")
			return
		}
		fmt.Println("Master Pointer p1 unit32: " + fmt.Sprint(testMemoryAddress))
		pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		fmt.Println("Master Pointer points to: " + fmt.Sprint(pointsTo))
		fmt.Println("Type of Object Pointed to:")
		fmt.Println(reflect.TypeOf(pointsTo))

		fmt.Println("Offset: 00000034")
		offset, offsetError = w32.HexToUint32("00000034")
		if offsetError != nil {
			//Error Getting Offset: -10
			fmt.Println("-10")
			return
		}
		fmt.Println("Offset as uint32: " + fmt.Sprint(offset))

		var objectPointer uint32

		objectPointer = pointsTo + offset
		fmt.Println("Object Location Pointer + offset: " + fmt.Sprint(objectPointer))
		objectLoc, err := w32.ReadProcessMemoryAsUint32(handle, objectPointer)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		}
		fmt.Println("Object Location: " + fmt.Sprint(objectLoc))

		fmt.Println("Object Data Offset Hex: 00000014")
		var objectDataOffset uint32
		var objectDataOffsetError error
		objectDataOffset, objectDataOffsetError = w32.HexToUint32("00000014")
		if objectDataOffsetError != nil {
			//Error setting Object Offset: -11
			fmt.Println("-11")
			return
		}
		fmt.Println("Object Data Offset uint32: " + fmt.Sprint(objectDataOffset))
		var objectDataLocation uint32
		objectDataLocation = objectLoc + objectDataOffset
		fmt.Println("Object Data Location uint32: " + fmt.Sprint(objectDataLocation))

		data, err := w32.ReadProcessMemory(handle, objectDataLocation, uint(8))
		var spot float64
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		} else {
			fmt.Println("X: ")
			fmt.Println(reflect.TypeOf(data))
			fmt.Println(data)
			fmt.Println(string(data))
			spot = Float64frombytes(data)
			fmt.Println(Float64frombytes(data))
		}
		spot = spot + (1 * (math.Pow(10, 5)))
		fmt.Println(string(data))
		err = w32.WriteProcessMemory(handle, objectDataLocation, Float64bytes(spot), uint(8))
		if err != nil {
			fmt.Println("Err: " + err.Error())
		} else {
			fmt.Println("Success")
		}

		/*var character uint32 = 2
		err = w32.WriteProcessMemoryAsUint32(handle, objectDataLocation, character)
		if err != nil {
			fmt.Println("Err: " + err.Error())
		} else {
			fmt.Println("Success")
		}*/

	case "-test":
		var testMemoryAddress uint32
		var testError error
		var address string = "01DEA6E0"

		for i := 0; i < 100; i++ {

			//fmt.Println("Master Pointer p1 HEX: " + address)
			testMemoryAddress, testError = w32.HexToUint32(address)
			if testError != nil {
				//Error getting Stage Stage Uint32: -3
				fmt.Println("-3")
				return
			}
			var tOff uint32 = uint32(i) * 4
			testMemoryAddress += tOff
			fmt.Println("Looking At Address: " + fmt.Sprint(testMemoryAddress))
			pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				return
			}
			fmt.Println(fmt.Sprint(i) + ": " + fmt.Sprint(pointsTo))
		}
	case "-test2":
		var testMemoryAddress uint32 = 124653080

		for i := 0; i < 50; i++ {

			var tOff uint32 = uint32(i) * 4
			testMemoryAddress += tOff
			fmt.Println("Looking At Address: " + fmt.Sprint(testMemoryAddress))
			/*pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				return
			}
			fmt.Println(fmt.Sprint(i) + ": " + fmt.Sprint(pointsTo))*/
			data, err := w32.ReadProcessMemory(handle, testMemoryAddress, uint(32))
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				fmt.Println(err)

			} else {
				fmt.Println(string(data))
				//fmt.Println(strconv.Itoa(data))

			}
		}
	case "-check":
		var testMemoryAddress uint32 = 124652088

		fmt.Println("Looking At Address: " + fmt.Sprint(testMemoryAddress))
		pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		} else {
			fmt.Println(pointsTo)
		}
		data, err := w32.ReadProcessMemory(handle, testMemoryAddress, uint(2048))
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)

		} else {
			fmt.Println(string(data))
			//fmt.Println(strconv.Itoa(data))

		}
	case "-list":
		var testMemoryAddress uint32
		var testError error

		fmt.Println("Master Pointer p1 HEX: 01DEA6E0")
		testMemoryAddress, testError = w32.HexToUint32("01DEA6E0")
		if testError != nil {
			//Error getting Stage Stage Uint32: -3
			fmt.Println("-3")
			return
		}
		fmt.Println("Master Pointer p1 unit32: " + fmt.Sprint(testMemoryAddress))
		pointsTo, err := w32.ReadProcessMemoryAsUint32(handle, testMemoryAddress)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		fmt.Println("Master Pointer points to: " + fmt.Sprint(pointsTo))
		//fmt.Println("Type of Object Pointed to:")
		//fmt.Println(reflect.TypeOf(pointsTo))
		findName(pointsTo, handle)
		getXValue(pointsTo, handle)
		getYValue(pointsTo, handle)
		getZValue(pointsTo, handle)

		var testRing uint32 = 112956268
		getXValue(testRing, handle)
		getYValue(testRing, handle)
		getZValue(testRing, handle)
		var testX float32 = 227.8272
		var testY float32 = 70
		var testZ float32 = -414.98663
		setXValue(testRing, handle, testX)
		setYValue(testRing, handle, testY)
		setZValue(testRing, handle, testZ)
		//	getXValue(testRing, handle)
		//	getYValue(testRing, handle)
		//	getZValue(testRing, handle)
		//setXValue(pointsTo, handle, 140)

		/*fmt.Println("Offset: 00000044")

		offset, offsetError = w32.HexToUint32("00000044")
		if offsetError != nil {
			//Error Getting Offset: -10
			fmt.Println("-10")
			return
		}
		fmt.Println("Offset as uint32: " + fmt.Sprint(offset))

		var objectPointer uint32

		objectPointer = pointsTo + offset
		fmt.Println("Object Location Pointer + offset: " + fmt.Sprint(objectPointer))
		objectLoc, err := w32.ReadProcessMemoryAsUint32(handle, objectPointer)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		}
		fmt.Println("Object Location: " + fmt.Sprint(objectLoc))

		fmt.Println("Object Data Offset Hex: 00000000")
		var objectDataOffset uint32
		var objectDataOffsetError error
		objectDataOffset, objectDataOffsetError = w32.HexToUint32("00000000")
		if objectDataOffsetError != nil {
			//Error setting Object Offset: -11
			fmt.Println("-11")
			return
		}
		fmt.Println("Object Data Offset uint32: " + fmt.Sprint(objectDataOffset))
		var objectDataLocation uint32
		objectDataLocation = objectLoc + objectDataOffset
		fmt.Println("Object Data Location uint32: " + fmt.Sprint(objectDataLocation))

		data, err := w32.ReadProcessMemory(handle, objectDataLocation, uint(16))
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(data))
			//fmt.Println(strconv.Itoa(data))

		}
		*/

		var seenObjects [2000]uint32
		var firstSeen uint32
		firstSeen = pointsTo
		var seenCount int
		for i := 1; i < 2000; i++ {
			//fmt.Println("Next Object Offest: 4")
			var nextObjectOffset uint32
			var nextObjectOffsetError error
			nextObjectOffset, nextObjectOffsetError = w32.HexToUint32("00000000")
			if nextObjectOffsetError != nil {
				//Error Getting next object  Offset: -40
				fmt.Println("-40")
				return
			}
			var nextobjectPointer uint32

			nextobjectPointer = pointsTo + nextObjectOffset
			//fmt.Println("Next Object pointer: " + fmt.Sprint(nextobjectPointer))

			nextObjectLoc, err := w32.ReadProcessMemoryAsUint32(handle, nextobjectPointer)
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				fmt.Println(err)
				return
			}
			//fmt.Println("Next Object Location: " + fmt.Sprint(nextObjectLoc))
			//	for j := 1; j < i; j++ {
			if firstSeen == nextObjectLoc {
				fmt.Println("Seen Before: " + fmt.Sprint(nextObjectLoc))
				seenCount++
				if seenCount >= 1 {

					os.Exit(0)
				}

				//	}
			}
			seenObjects[i] = nextObjectLoc
			name := findName(nextObjectLoc, handle)
			if checkIfNew(name, oldObjects) {
				fmt.Print(fmt.Sprint(i) + ": ")
				fmt.Println(name)
				fmt.Println(nextObjectLoc)
			}

			pointsTo = nextObjectLoc
			//getXValue(nextObjectLoc, handle)

		}

		/*
			var nextObjectNameOffset uint32
			var nextObjectNameOffsetError error
			var nextObjectNameLoc uint32
			nextObjectNameOffset, nextObjectNameOffsetError = w32.HexToUint32("00000044")
			if nextObjectNameOffsetError != nil {
				fmt.Println(nextObjectNameOffsetError)
			}
			nextObjectNameLoc = nextObjectLoc + nextObjectNameOffset
			nextObjectName, err := w32.ReadProcessMemory(handle, nextObjectNameLoc, uint(32))
			if err != nil {
				//Error Reading Memory: -1
				fmt.Println("-1")
				fmt.Println(err)
				return
			} else {
				fmt.Println(reflect.TypeOf(nextObjectName))
				fmt.Println(string(nextObjectName))
				//fmt.Println(strconv.Itoa(data))

			}
		*/

	case "-findCamera":
		var correctCount int = 0
		globalPointerToPlayer := uint32(0x01DEA6E0)
		playerAddress, err := w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		var xPos = getX(playerAddress, handle)
		var yPos = getY(playerAddress, handle)
		var zPos = getZ(playerAddress, handle)
		var seenX float32
		var seenY float32
		var seenZ float32
		//data1Offset := 0x00000034
		//data1Pointer := playerAddress + uint32(data1Offset)
		//data1Location, err := w32.ReadProcessMemoryAsUint32(handle, data1Pointer)

		if err != nil {
			//Error Reading Memory: -1
			fmt.Println("-1")
			return
		}
		fmt.Println(xPos)
		fmt.Println(yPos)
		fmt.Println(zPos)
		locations := make([]uint32, 0)
		//return
		var checkRange float32 = 100
		var startSpot uint32 = 30000000
		for offset := 0; offset < 10000000; offset++ {
			if offset%1000000 == 0 {
				fmt.Print("Offset: ")
				fmt.Println(offset)
			}
			value, err := w32.ReadProcessMemory(handle, startSpot+uint32(offset), uint(4))
			if err != nil {
				//Error Reading Memory: -1
				//fmt.Println("-200: Error Reading Object 2 Data")
				//fmt.Println(err)
				//return
			} else {
				valueFloat := Float32frombytes(value)
				if correctCount == 0 {
					if valueFloat > (xPos-checkRange) && valueFloat < (xPos+checkRange) {
						seenX = valueFloat
						correctCount = 1
						offset = offset + 3
					}
				} else if correctCount == 1 {
					if valueFloat > (yPos-checkRange) && valueFloat < (yPos+checkRange) {
						seenY = valueFloat
						correctCount = 2
						offset = offset + 3
					} else {
						correctCount = 0
					}
				} else if correctCount == 2 {
					if valueFloat > (zPos-checkRange) && valueFloat < (zPos+checkRange) {
						seenZ = valueFloat
						fmt.Print("Start of Data Section: ")
						fmt.Println(startSpot + uint32(offset-(8+0x14)))
						locations = append(locations, startSpot+uint32(offset-(8+0x14)))
						fmt.Print("X Pos: ")
						fmt.Print(startSpot + uint32(offset-8))
						fmt.Print(", Value: ")
						fmt.Println(seenX)

						fmt.Print("Y Pos: ")
						fmt.Print(startSpot + uint32(offset-4))
						fmt.Print(", Value: ")
						fmt.Println(seenY)
						fmt.Print("Z Pos: ")
						fmt.Print(startSpot + uint32(offset))
						fmt.Print(", Value: ")
						fmt.Println(seenZ)
					} else {
						correctCount = 0
					}
				}
			}
		}
		for i := 0; i < len(locations); i++ {
			buf := bufio.NewReader(os.Stdin)
			fmt.Print(locations[i])
			fmt.Print("> ")
			_, err := buf.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
			} else {
				err2 := w32.WriteProcessMemory(handle, locations[i]+0x14, []byte{0}, uint(8))
				if err2 != nil {
					//Error Reading Memory: -1
					fmt.Println("-200: Error Reading Object 2 Data")
					fmt.Println(err)
					return
				} else {
					//fmt.Print("Location Set")
					//fmt.Println(Float32frombytes(name))
					//fmt.Println(strconv.Itoa(data))

				}
			}
		}

	case "-change":
		buf := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		var i int
		_, _ = fmt.Scanf("%d", &i)
		var spotLoc = uint32(i)
		fmt.Println("Setting 0x14, x?")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0x14, []byte{0}, uint(4))

		fmt.Println("Setting 0x18, y?")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0x18, []byte{0}, uint(4))

		fmt.Println("Setting 0x1C, z?")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0x1C, []byte{0}, uint(4))

		fmt.Println("Setting 0x8, x rot??")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0x8, []byte{0}, uint(4))

		fmt.Println("Setting 0xC, y rot??")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0xC, []byte{0}, uint(4))

		fmt.Println("Setting 0x10, z rot??")
		fmt.Print("> ")
		buf.ReadBytes('\n')
		_ = w32.WriteProcessMemory(handle, spotLoc+0x10, []byte{0}, uint(4))

		/*
			fmt.Println("Setting 0x0, ?")
			fmt.Print("> ")
			buf.ReadBytes('\n')
			_ = w32.WriteProcessMemory(handle, spotLoc+0x0, []byte{0}, uint(4))
			fmt.Println("Setting 0x4, ?")
			fmt.Print("> ")
			buf.ReadBytes('\n')
			_ = w32.WriteProcessMemory(handle, spotLoc+0x4, []byte{0}, uint(4))

			fmt.Println("Setting 0x20, ?")
			fmt.Print("> ")
			buf.ReadBytes('\n')
			_ = w32.WriteProcessMemory(handle, spotLoc+0x20, []byte{0}, uint(4))
			fmt.Println("Setting 0x24, ?")
			fmt.Print("> ")
			buf.ReadBytes('\n')
			_ = w32.WriteProcessMemory(handle, spotLoc+0x24, []byte{0}, uint(4))
		*/

	case "":

		//twoDSonic(playerAddress, handle, 0)

		xinput.Load()
		fmt.Println("Keep me open")
		w32.MessageBox(0, "Special Thanks to Jelly for help. \n Only xbox 360 controller supported \nL Bumper = Save State, R Bumper = Load State\nComing soon: Save gravity (like in CG)\nSave Time And Rings\nSave Camera\nSave Hourglass State\nSave Powerup\nMake you undead if you are dead\nSuggestions?\nBuilt @ 9:37 PM EST, 25/2/2017", "SPaDeS-Tools", 0)
		/*

			devs, err := keylogger.NewDevices()
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, val := range devs {
				fmt.Println("Id->", val.Id, "Device->", val.Name)
			}

			fmt.Print("which device (keyboard?)> ")
			var i int
			_, _ = fmt.Scanf("%d", &i)
			rd := keylogger.NewKeyLogger(devs[i])
			in, _ := rd.Read()
			for i := range in {
				//listen only key stroke event
				if i.Type == keylogger.EV_KEY {
					fmt.Println(i.KeyString())
				}
			}*/
		var scale []byte
		globalPointerToPlayer, _ := w32.HexToUint32("01DEA6E0")

		var playerAddress uint32
		playerAddress, _ = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)

		scale = getScale(playerAddress, handle)
		var outTE *walk.TextEdit
		window :=
			MainWindow{
				Title:   "SPaDeS SA2 Tool(s?)",
				MinSize: Size{600, 400},
				Layout:  VBox{},
				Children: []Widget{

					PushButton{
						Text: "Start Save Engine",
						OnClicked: func() {
							line4 := "More Features / Updates Coming Soon(ish)*Maybe\r\n"
							line5 := "by http://twitch.tv/spadespwnzyou\r\n"
							line6 := "https://discord.gg/Ge3gtuZ to suggest updates\r\n"
							outTE.SetText("Save System Enabled\r\n Left bumper (360 controller) or Left Arrow to save state\r\nRight Bumper or Right Arrow to Load State\r\n" + line4 + line5 + line6)
							go saveEngine()
							go keyboardThread()
						},
					},
					PushButton{
						Text: "99 Lives",
						OnClicked: func() {
							set99Lives(handle)
						},
					},
					PushButton{
						Text: "Reset Flatness",
						OnClicked: func() {
							playerAddress, _ = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
							setScale(playerAddress, handle, scale)
						},
					},
					PushButton{
						Text: "Flat X",
						OnClicked: func() {
							playerAddress, _ = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
							flatX(playerAddress, handle)
						},
					},
					PushButton{
						Text: "Flat Y ( shadow% races when?)",
						OnClicked: func() {
							playerAddress, _ = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
							flatY(playerAddress, handle)
						},
					},
					PushButton{
						Text: "Flat Z",
						OnClicked: func() {
							playerAddress, _ = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
							flatZ(playerAddress, handle)
						},
					},
					HSplitter{
						Children: []Widget{
							TextEdit{AssignTo: &outTE, ReadOnly: true},
						},
					},
				},
			}
		window.Run()
		//go saveEngine()

	}

}

func checkByte(b byte, spot int) byte {
	spotUint := uint8(spot)
	spotByte := byte(spotUint)
	return b & spotByte
}

func saveEngine() {
	var globalPointerToPlayer uint32
	var testError error
	var x float32
	var y float32
	var z float32
	var horSpeed float32
	var vertSpeed float32
	var xRot []byte
	var yRot []byte
	var zRot []byte
	var status []byte
	var hangtime []byte
	var physics []byte
	var animFrames []byte
	var currAnim []byte
	var action []byte
	var hover []byte
	var momentum float32
	var savingState bool = false
	var loadingState bool = false
	//	var cam1X float32
	//	var cam1Y float32
	//	var cam1Z float32
	var cam2 []byte
	var gametime []byte
	var controllerState xinput.State
	globalPointerToPlayer, testError = w32.HexToUint32("01DEA6E0")
	if testError != nil {
		fmt.Println("-3")
		return
	}
	var playerAddress uint32
	var gravity []byte
	playerAddress, err := w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
	if err != nil {
		//Error Reading Memory: -1
		fmt.Println("-1")
		return
	}
	for {
		xinput.GetState(1, &controllerState)
		//fmt.Println("State: " + strconv.Itoa(int(controllerState.Gamepad.Buttons)))

		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(controllerState.Gamepad.Buttons))
		if savingState == true {
			if checkByte(b[1], 1) == 0 {
				savingState = false

				fmt.Println("Done Saving")
			}
		}
		if savingState == false {
			if checkByte(b[1], 1) > 0 || savestateKey == true {
				savestateKey = false
				x = getX(playerAddress, handle)
				y = getY(playerAddress, handle)
				z = getZ(playerAddress, handle)
				horSpeed = getHorSpeed(playerAddress, handle)
				vertSpeed = getVertSpeed(playerAddress, handle)
				xRot = getXRot(playerAddress, handle)
				yRot = getYRot(playerAddress, handle)
				zRot = getZRot(playerAddress, handle)
				status = getStatus(playerAddress, handle)
				hangtime = getHangTime(playerAddress, handle)
				physics = getPhysics(playerAddress, handle)
				animFrames = getAnimFrames(playerAddress, handle)
				currAnim = getCurrAnim(playerAddress, handle)
				action = getAction(playerAddress, handle)
				hover = getHover(handle)
				momentum = getMomentum(handle)
				gametime = getTime(handle)
				//	cam1X = getCamera1X(handle)
				//	cam1Y = getCamera1Y(handle)
				//	cam1Z = getCamera1Z(handle)
				cam2 = getCamera2(handle)
				gravity = getGravity(handle)
				if x == 0 && y == 0 && z == 0 {
					playerAddress, err = w32.ReadProcessMemoryAsUint32(handle, globalPointerToPlayer)
					if err != nil {
						//Error Reading Memory: -1
						fmt.Println("-1")
						return
					}
					cam2 = getCamera2(handle)
					x = getX(playerAddress, handle)
					y = getY(playerAddress, handle)
					z = getZ(playerAddress, handle)
					horSpeed = getHorSpeed(playerAddress, handle)
					vertSpeed = getVertSpeed(playerAddress, handle)
					xRot = getXRot(playerAddress, handle)
					yRot = getYRot(playerAddress, handle)
					zRot = getZRot(playerAddress, handle)
					status = getStatus(playerAddress, handle)
					hangtime = getHangTime(playerAddress, handle)
					physics = getPhysics(playerAddress, handle)
					animFrames = getAnimFrames(playerAddress, handle)
					currAnim = getCurrAnim(playerAddress, handle)
					action = getAction(playerAddress, handle)
					hover = getHover(handle)
					momentum = getMomentum(handle)
					gametime = getTime(handle)
					gravity = getGravity(handle)
					//	cam1X = getCamera1X(handle)
					//	cam1Y = getCamera1Y(handle)
					//	cam1Z = getCamera1Z(handle)
					//cam2 = getCamera2(handle)

				}
				savingState = true

				fmt.Println("Saving State")
			}
		}

		if loadingState == true {
			if checkByte(b[1], 2) == 0 {
				loadingState = false

				fmt.Println("Done Loading")
			}
		}
		if loadingState == false {
			if checkByte(b[1], 2) > 0 || loadsateKey == true {
				loadsateKey = false
				loadingState = true
				//setCamera2(handle, cam2)

				setXValue(playerAddress, handle, x)
				setYValue(playerAddress, handle, y)
				setZValue(playerAddress, handle, z)
				setHorSpeed(playerAddress, handle, horSpeed)
				setVertSpeed(playerAddress, handle, vertSpeed)
				setXRot(playerAddress, handle, xRot)
				setYRot(playerAddress, handle, yRot)
				setZRot(playerAddress, handle, zRot)
				setStatus(playerAddress, handle, status)
				setHangTime(playerAddress, handle, hangtime)
				setPhysics(playerAddress, handle, physics)
				setAnimFrames(playerAddress, handle, animFrames)
				setCurrAnim(playerAddress, handle, currAnim)
				setAction(playerAddress, handle, action)
				setHover(handle, hover)
				setMomentum(handle, momentum)
				setTime(handle, gametime)
				setCamera2(handle, cam2)
				setGravity(handle, gravity)

				time.Sleep(time.Millisecond * 100)

				setXValue(playerAddress, handle, x)
				setYValue(playerAddress, handle, y)
				setZValue(playerAddress, handle, z)
				setHorSpeed(playerAddress, handle, horSpeed)
				setVertSpeed(playerAddress, handle, vertSpeed)
				setXRot(playerAddress, handle, xRot)
				setYRot(playerAddress, handle, yRot)
				setZRot(playerAddress, handle, zRot)
				setStatus(playerAddress, handle, status)
				setHangTime(playerAddress, handle, hangtime)
				setPhysics(playerAddress, handle, physics)
				setAnimFrames(playerAddress, handle, animFrames)
				setCurrAnim(playerAddress, handle, currAnim)
				setAction(playerAddress, handle, action)
				setHover(handle, hover)
				setMomentum(handle, momentum)
				setTime(handle, gametime)
				setCamera2(handle, cam2)
				setGravity(handle, gravity)
				fmt.Println("Loading State")
			}
		}
	}
}
