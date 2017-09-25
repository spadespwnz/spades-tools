package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tajtiattila/xinput"
	"github.com/vence722/inputhook"
	"github.com/xackery/w32"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type State struct {
	ready      bool
	level      int
	x          float32
	y          float32
	z          float32
	horSpeed   float32
	vertSpeed  float32
	xRot       []byte
	yRot       []byte
	zRot       []byte
	status     []byte
	hangtime   []byte
	physics    []byte
	animFrames []byte
	currAnim   []byte
	action     []byte
	hover      []byte
	momentum   float32
	cam2       []byte
	gametime   []byte
	gravity    []byte
}

type Racer struct {
	playerID    int
	pointerData []byte
}
type Pack struct {
	Pack  string `json:"pack"`
	Times []int  `json:"times"`
}
type Result struct {
	Times []int
}
type GhostData struct {
	Obj1 []byte
	Obj2 []byte
}

type LoginResponse struct {
	Login string `json:"login"`
	Error string `json:"error"`
	Token string `json:"token"`
}

var currPack string
var currTimes Result
var speedrunTimes Result
var raceTimes Result

type PlayerData struct {
	obj1 []byte
	obj2 []byte
	grav []byte
	time []byte
}

type Obj2Sizes struct {
	SONIC    uint
	TALES    uint
	KNUCKLES uint
	EGGMAN   uint
	SHADOW   uint
	ROUGE    uint
}

type MyMainWindow struct {
	*walk.MainWindow
}

var db *sql.DB
var obj2Sizes Obj2Sizes
var savestateKey = false
var loadsateKey = false
var stopListening = true
var handle w32.HANDLE
var oldPointerData []uint8
var Obj2Size uint = 0x20
var dataWriteLock = false
var serverButton *walk.PushButton
var connectButton *walk.PushButton

var raceAgainst *walk.LineEdit
var raceAgainstValues []string
var raceAgainstSel = 0

var saveTo *walk.LineEdit
var saveToValues []string
var saveToSel = 0

var drawLoc uint32 = 0x130C000
var tempSonicObj1Data []byte
var tempSonicObj2Data []byte
var tempSonicPointerData []byte
var obj1Loc uint32 = 0x1300050
var obj2Loc uint32 = 0x1300100
var insertLoc uint32 = 0x1300000
var ghostMod = false
var showGhost = true
var saveGhost = false
var speedrunRace = false

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
		fmt.Printf("Process: %s, ID: %d\n", GetProcessName(pid), pid)
		if strings.Contains(GetProcessName(pid), name) {
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

func checkIfNew(item string, seenList [1]string) bool {
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
func simpleCopyToOffset(offset int) {

	location, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
	var obj2CopySize = uint(0)
	currChar, _ := w32.ReadProcessMemory(handle, 0x1934B80, uint(1))
	var charInt = int(currChar[0])

	if charInt == 0 {
		obj2CopySize = obj2Sizes.SONIC
	} else if charInt == 6 {
		obj2CopySize = obj2Sizes.TALES
	} else if charInt == 4 {
		obj2CopySize = obj2Sizes.KNUCKLES
	} else if charInt == 7 {
		obj2CopySize = obj2Sizes.EGGMAN
	} else if charInt == 5 {
		obj2CopySize = obj2Sizes.ROUGE
	} else if charInt == 1 {
		obj2CopySize = obj2Sizes.SHADOW
	}
	//level, _ := w32.ReadProcessMemory(handle, 0x01934B70, uint(1))
	//levelInt := int(level[0])

	if charInt == 4 || charInt == 5 {
		copyKnuxToOffset(offset)
		return

	}

	var objOrigData1Loc uint32
	var objOrigData2Loc uint32

	masterObj, _ := w32.ReadProcessMemory(handle, location, 0x50)
	masterCopy := make([]uint8, 0x50)
	copy(masterCopy, masterObj)
	tempLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempLoc, location)
	copy(masterCopy[0x0:0x04], tempLoc)

	copy(masterCopy[0x10:0x14], []uint8{0, 0, 0, 0})

	var someLen = (0x34 - 0x18)
	someBytes := make([]uint8, someLen)
	for i := 0; i < someLen; i++ {
		someBytes[i] = 0
	}
	copy(masterCopy[0x18:0x34], someBytes)

	copy(masterCopy[0x44:0x50], []uint8{0, 0, 0, 0, 0, 0})

	//Set Data Storage locations

	temp1Loc := make([]uint8, 4)
	temp2Loc := make([]uint8, 4)

	binary.LittleEndian.PutUint32(temp1Loc, obj1Loc+uint32(offset*0x1000))
	binary.LittleEndian.PutUint32(temp2Loc, obj2Loc+uint32(offset*0x1000))

	copy(masterCopy[0x34:0x38], temp1Loc)

	copy(masterCopy[0x40:0x44], temp2Loc)

	//Copy Data into locations
	objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)

	objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)

	obj1Data, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, 0x30)

	obj2Data, _ := w32.ReadProcessMemory(handle, objOrigData2Loc, obj2CopySize)

	w32.WriteProcessMemory(handle, obj1Loc+uint32(offset*0x1000), obj1Data, 0x30)

	w32.WriteProcessMemory(handle, obj2Loc+uint32(offset*0x1000), obj2Data, obj2CopySize)
	w32.WriteProcessMemory(handle, insertLoc+uint32(offset*0x1000), masterCopy, 0x50)

	nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, location+0x04)
	w32.WriteProcessMemoryAsUint32(handle, location+0x04, insertLoc+uint32(offset*0x1000))
	w32.WriteProcessMemoryAsUint32(handle, nextAddr, insertLoc+uint32(offset*0x1000))

	var copyLen = uint(0x14 - 0x8)
	allOverWriteData := make([]uint8, copyLen)
	for i := 0; i < len(allOverWriteData); i++ {
		allOverWriteData[i] = 0
	}

	w32.WriteProcessMemory(handle, insertLoc+uint32(offset*0x1000)+0x8, allOverWriteData, copyLen)

}
func simpleCopy(location uint32, handle w32.HANDLE) {

	fmt.Printf("Simple Copy using %s %s\n", location, handle)
	var obj2CopySize = uint(0)
	currChar, _ := w32.ReadProcessMemory(handle, 0x1934B80, uint(1))
	var charInt = int(currChar[0])
	fmt.Printf("Using Char: %d\n", charInt)
	if charInt == 0 {
		obj2CopySize = obj2Sizes.SONIC
	} else if charInt == 6 {
		obj2CopySize = obj2Sizes.TALES
	} else if charInt == 4 {
		obj2CopySize = obj2Sizes.KNUCKLES
	} else if charInt == 7 {
		obj2CopySize = obj2Sizes.EGGMAN
	} else if charInt == 5 {
		obj2CopySize = obj2Sizes.ROUGE
	} else if charInt == 1 {
		obj2CopySize = obj2Sizes.SHADOW
	}
	level, _ := w32.ReadProcessMemory(handle, 0x01934B70, uint(1))
	levelInt := int(level[0])
	fmt.Printf("Level: %d\n", levelInt)

	if charInt == 4 || charInt == 5 {
		copyKnux()
		return

	}

	var objOrigData1Loc uint32
	var objOrigData2Loc uint32

	masterObj, _ := w32.ReadProcessMemory(handle, location, 0x50)
	masterCopy := make([]uint8, 0x50)
	copy(masterCopy, masterObj)
	tempLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempLoc, location)
	copy(masterCopy[0x0:0x04], tempLoc)

	copy(masterCopy[0x10:0x14], []uint8{0, 0, 0, 0})

	var someLen = (0x34 - 0x18)
	someBytes := make([]uint8, someLen)
	for i := 0; i < someLen; i++ {
		someBytes[i] = 0
	}
	copy(masterCopy[0x18:0x34], someBytes)

	copy(masterCopy[0x44:0x50], []uint8{0, 0, 0, 0, 0, 0})

	//Set Data Storage locations

	temp1Loc := make([]uint8, 4)
	temp2Loc := make([]uint8, 4)

	binary.LittleEndian.PutUint32(temp1Loc, obj1Loc)
	binary.LittleEndian.PutUint32(temp2Loc, obj2Loc)

	copy(masterCopy[0x34:0x38], temp1Loc)

	copy(masterCopy[0x40:0x44], temp2Loc)

	//Copy Data into locations
	objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)

	objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)

	obj1Data, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, 0x30)

	fmt.Printf("Copy Size: %d\n", obj2CopySize)
	obj2Data, _ := w32.ReadProcessMemory(handle, objOrigData2Loc, obj2CopySize)

	w32.WriteProcessMemory(handle, obj1Loc, obj1Data, 0x30)

	w32.WriteProcessMemory(handle, obj2Loc, obj2Data, obj2CopySize)
	w32.WriteProcessMemory(handle, insertLoc, masterCopy, 0x50)

	nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, location+0x04)
	w32.WriteProcessMemoryAsUint32(handle, location+0x04, insertLoc)
	w32.WriteProcessMemoryAsUint32(handle, nextAddr, insertLoc)

	var copyLen = uint(0x14 - 0x8)
	allOverWriteData := make([]uint8, copyLen)
	for i := 0; i < len(allOverWriteData); i++ {
		allOverWriteData[i] = 0
	}

	w32.WriteProcessMemory(handle, insertLoc+0x8, allOverWriteData, copyLen)
	oldPointerData, _ = w32.ReadProcessMemory(handle, insertLoc, uint(0x50))

}

func copyKnux() {
	pointerData := make([]uint8, 0x50)
	obj1Data := make([]uint8, 0x30)

	tempDisplayLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempDisplayLoc, 0x6C8280)
	copy(pointerData[0x14:0x18], tempDisplayLoc)

	temp1Loc := make([]uint8, 4)

	binary.LittleEndian.PutUint32(temp1Loc, obj1Loc)

	copy(pointerData[0x34:0x38], temp1Loc)

	playerLoc, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
	playerObj1, _ := w32.ReadProcessMemoryAsUint32(handle, playerLoc+0x34)
	locData, _ := w32.ReadProcessMemory(handle, playerObj1+0x14, 0x20-0x14)

	var float1 float32 = 1.0
	bits := math.Float32bits(float1)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	copy(obj1Data[0x28:0x2C], Float32bytes(float1))
	w32.WriteProcessMemory(handle, obj1Loc, obj1Data, 0x30)

	w32.WriteProcessMemory(handle, obj1Loc+0x14, locData, 0x20-0x14)
	location, _ := w32.ReadProcessMemoryAsUint32(handle, 0x1A5A25C)
	p, _ := w32.ReadProcessMemory(handle, location, 0x50)
	copy(pointerData[0x0:0x8], p[0x0:0x8])

	tempLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempLoc, location)
	copy(pointerData[0x0:0x04], tempLoc)

	w32.WriteProcessMemory(handle, insertLoc, pointerData, 0x50)

	nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, location+0x04)
	w32.WriteProcessMemoryAsUint32(handle, location+0x04, insertLoc)
	w32.WriteProcessMemoryAsUint32(handle, nextAddr, insertLoc)

	oldPointerData, _ = w32.ReadProcessMemory(handle, insertLoc, uint(0x50))

}
func copyKnuxToOffset(offset int) {
	pointerData := make([]uint8, 0x50)
	obj1Data := make([]uint8, 0x30)

	tempDisplayLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempDisplayLoc, 0x6C8280)
	copy(pointerData[0x14:0x18], tempDisplayLoc)

	temp1Loc := make([]uint8, 4)

	binary.LittleEndian.PutUint32(temp1Loc, obj1Loc+uint32(offset*0x1000))

	copy(pointerData[0x34:0x38], temp1Loc)

	playerLoc, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
	playerObj1, _ := w32.ReadProcessMemoryAsUint32(handle, playerLoc+0x34)
	locData, _ := w32.ReadProcessMemory(handle, playerObj1+0x14, 0x20-0x14)

	var float1 float32 = 1.0
	bits := math.Float32bits(float1)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	copy(obj1Data[0x28:0x2C], Float32bytes(float1))
	w32.WriteProcessMemory(handle, obj1Loc+uint32(offset*0x1000), obj1Data, 0x30)

	w32.WriteProcessMemory(handle, obj1Loc+uint32(offset*0x1000)+0x14, locData, 0x20-0x14)
	location, _ := w32.ReadProcessMemoryAsUint32(handle, 0x1A5A25C)
	p, _ := w32.ReadProcessMemory(handle, location, 0x50)
	copy(pointerData[0x0:0x8], p[0x0:0x8])

	tempLoc := make([]uint8, 4)
	binary.LittleEndian.PutUint32(tempLoc, location)
	copy(pointerData[0x0:0x04], tempLoc)

	w32.WriteProcessMemory(handle, insertLoc+uint32(offset*0x1000), pointerData, 0x50)

	nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, location+0x04)
	w32.WriteProcessMemoryAsUint32(handle, location+0x04, insertLoc+uint32(offset*0x1000))
	w32.WriteProcessMemoryAsUint32(handle, nextAddr, insertLoc+uint32(offset*0x1000))

	oldPointerData, _ = w32.ReadProcessMemory(handle, insertLoc+uint32(offset*0x1000), uint(0x50))
}

func copyPlayer(location uint32, handle w32.HANDLE) {

	var objOrigData1Loc uint32
	var objOrigData2Loc uint32

	var lastMenu int = -1
	/*
		0 = menus
		1 = loading level
		2 = loading level
		7 = finished loading (happens on restart)
		14 = game over?
		16 = game running
		17 = pause
		21 = exit

	*/
	var replayData = make(map[string]PlayerData)

	var recording = true
	var oldMili = -1

	var firstFullLoad = true

	for {

		menu, _ := w32.ReadProcessMemory(handle, 0x01934BE0, 1)
		menuItem := menu[0]
		menuInt := int(menuItem)
		if menuInt != lastMenu {
			fmt.Printf("loc: %d\n", location)
			fmt.Printf("Menu: %d\n", menuInt)
			if menuInt == 17 {
				secondObjLoc, _ := w32.ReadProcessMemoryAsUint32(handle, insertLoc+0x4)
				w32.WriteProcessMemoryAsUint32(handle, location+0x4, secondObjLoc)
				w32.WriteProcessMemoryAsUint32(handle, secondObjLoc, location)

			}
			if menuInt == 14 {
				secondObjLoc, _ := w32.ReadProcessMemoryAsUint32(handle, insertLoc+0x4)
				w32.WriteProcessMemoryAsUint32(handle, location+0x4, secondObjLoc)
				w32.WriteProcessMemoryAsUint32(handle, secondObjLoc, location)

			}
			if menuInt == 16 {
				if firstFullLoad {
					firstFullLoad = false
					fmt.Printf("Recording: %s\n", recording)
					if recording {

					} else {
						fmt.Printf("Creating Second Sonic\n")

						simpleCopy(location, handle)

					}
				} else {
					fmt.Printf("reloading character\n")
					nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, location+0x04)
					w32.WriteProcessMemoryAsUint32(handle, location+0x04, insertLoc)
					w32.WriteProcessMemoryAsUint32(handle, nextAddr, insertLoc)
				}
			}
			if menuInt == 7 {
				location, _ = w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
				firstFullLoad = true
				fmt.Printf("here 1\n")
				objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)
				objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)
				fmt.Printf("here 2\n")
				if recording {
					fmt.Printf("here 3\n")
					recording = false
				} else {
					recording = true
					replayData = make(map[string]PlayerData)

				}
			}

		}
		lastMenu = menuInt
		if menuInt == 16 {
			currTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, 3)
			min := int(currTime[0])
			sec := int(currTime[1])
			mili := int(currTime[2])
			var timeAsString = string(min) + string(sec) + string(mili)
			if mili != oldMili {
				if recording {
					fmt.Printf("recording\n")
					tempData := new(PlayerData)
					data, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, 0x30)
					tempData.obj1 = data
					data, _ = w32.ReadProcessMemory(handle, objOrigData2Loc+0x174, Obj2Size)
					tempData.obj2 = data
					replayData[timeAsString] = *tempData
				} else {
					if val, ok := replayData[timeAsString]; ok {
						fmt.Printf("Updating\n")
						w32.WriteProcessMemory(handle, obj1Loc, val.obj1, 0x30)
						w32.WriteProcessMemory(handle, obj2Loc+0x174, val.obj2, Obj2Size)

					}
				}
				oldMili = mili
			}
		}

		time.Sleep(1000000)
	}

}
func testMemory(handle w32.HANDLE) {
	var startAddr uint32 = 0xF000000
	for i := 0; i < 100000; i += 4 {
		data, _ := w32.ReadProcessMemory(handle, startAddr+uint32(i), uint(4))
		fmt.Printf("Addr: %d, Value: %d\n", startAddr+uint32(i), data)
	}

}

func validCloneLevel(level int, char int) bool {

	if level == 36 {
		return false
	} else if level == 70 {
		return false
	}
	return true
}

func getHandle() {
	var sonicPID uint32
	var err error
	sonicPID, err = FindProcessByName("sonic")

	reader := bufio.NewReader(os.Stdin)

	text := ""
	for {
		if sonicPID == 0 {
			fmt.Printf("Sonic not found\n")
			text = "n"
		} else {
			fmt.Printf("Use Process: %s (y / n) ?\n", GetProcessName(sonicPID))
			text, _ = reader.ReadString('\n')
			text = strings.TrimSpace(text)
		}

		if text == "y" {
			handle, err = w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, sonicPID)
			if err != nil {
				fmt.Printf("Invalid Application: " + err.Error() + "\n")
				os.Exit(0)

			}
			break
		} else if text == "n" {
			fmt.Printf("Enter ID to use (from List above): ")
			text, _ = reader.ReadString('\n')
			text = strings.TrimSpace(text)
			sonicPID64, nil := strconv.ParseUint(text, 10, 32)
			sonicPID = uint32(sonicPID64)
			handle, err = w32.OpenProcess(w32.PROCESS_ALL_ACCESS, false, sonicPID)
			if err != nil {
				fmt.Printf("Invalid Application")
				os.Exit(0)
			}
			break
		} else if text == "exit" {

			os.Exit(0)
		} else {
			fmt.Printf("Invalid Entry\n")

		}
	}
}
func main() {

	//var state *syscall.Proc
	obj2Sizes = Obj2Sizes{SONIC: 0x400, TALES: 0x460, KNUCKLES: 0x500, SHADOW: 0x400, EGGMAN: 0x460, ROUGE: 0x500}
	args := os.Args[1:]
	var command string
	if len(args) != 0 {
		command = args[0]
	}

	getHandle()

	switch command {

	case "-simpleCopy":
		playerDataAddr, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
		simpleCopy(playerDataAddr, handle)
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

	case "":
		states = make([]State, 10)

		raceAgainstValues = []string{"Last Attempt", "Longest Attempt"}
		saveToValues = []string{""}
		fmt.Println("Keep me open")

		db, _ = sql.Open("sqlite3", "./ghost.db")
		defer db.Close()
		countReq, _ := db.Query("select* from sqlite_master where type='table' and name='packs'")

		if countReq.Next() {
			fmt.Printf("Exists\n")
			statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS packs (name TEXT, owner INTEGER, times TEXT)")
			statement.Exec()
			defer statement.Close()
		} else {
			fmt.Printf("Doesnt\n")
			statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS packs (name TEXT, owner INTEGER, times TEXT)")
			statement.Exec()

			statement, _ = db.Prepare("CREATE UNIQUE INDEX idx_packs ON packs (name)")
			statement.Exec()
			defer statement.Close()
			//createGhostPack("Current_Attempt")
		}
		countReq.Close()

		var connAddr *walk.LineEdit
		var toggleGhostMod *walk.PushButton
		var toggleReplayMode *walk.PushButton
		var toggleSpeedrunMode *walk.PushButton
		var toggleRewinder *walk.PushButton
		var newPackTextEdit *walk.LineEdit
		var savePackTextEdit *walk.LineEdit
		var saveToLeft *walk.PushButton
		var saveToRight *walk.PushButton
		var saveButton *walk.PushButton
		var stateButton *walk.PushButton
		stopGhostMod := make(chan bool, 1)
		loadPacks()
		if _, err :=
			(MainWindow{

				Title:   "SPaDeS SA2 Tool(s?)",
				MinSize: Size{600, 300},
				Layout:  VBox{},
				Children: []Widget{
					PushButton{
						Text: "Reget connection to sonic app",
						OnClicked: func() {
							getHandle()
						},
					},
					PushButton{
						AssignTo: &stateButton,
						Text:     "Save Engine [Off]",
						OnClicked: func() {
							toggleRewinder.SetText("Rewinder [OFF]")
							rewinderEnabled = false
							if saveStates {
								saveStates = false
								stateButton.SetText("Save Engine [Off]")
							} else {
								go saveEngine()
								stateButton.SetText("Save Engine [On]")
							}

						},
					},
					PushButton{
						Text: "99 Lives",
						OnClicked: func() {
							set99Lives(handle)
						},
					},

					/*PushButton{
						AssignTo: &serverButton,
						Text:     "Start Server",
						OnClicked: func() {
							go startServer()
						},
					},*/
					HSplitter{
						Children: []Widget{
							LineEdit{
								AssignTo: &connAddr,
							},
							PushButton{
								AssignTo: &connectButton,
								Text:     "Connect to Room",
								OnClicked: func() {
									go connect(connAddr.Text())
								},
							},
						},
					},
					PushButton{
						Text: "Instructions",
						OnClicked: func() {
							openInstructions()
						},
					},
					HSplitter{
						Children: []Widget{
							Label{
								Text: "Race Against",
							},
							CheckBox{
								CheckState: walk.CheckChecked,
							},
							PushButton{
								Text: "<-",
								OnClicked: func() {
									raceAgainstPrev()
								},
							},
							LineEdit{
								AssignTo: &raceAgainst,
								Text:     raceAgainstValues[0],
								Enabled:  false,
							},

							PushButton{
								Text: "->",
								OnClicked: func() {
									raceAgainstNext()

								},
							},
							PushButton{
								Text: "Create New Pack",
								OnClicked: func() {
									popupWindow :=
										MainWindow{

											Title:   "Create Pack",
											MinSize: Size{300, 200},
											Layout:  VBox{},
											Children: []Widget{
												PushButton{
													Text: "Create",
													OnClicked: func() {
														createGhostPack(newPackTextEdit.Text())
													},
												},
												LineEdit{
													AssignTo: &newPackTextEdit,
													Text:     "",
													Enabled:  true,
												},
											},
										}
									popupWindow.Run()

								},
							},
						},
					},
					HSplitter{
						Children: []Widget{
							Label{
								Text: "Save To",
							},
							CheckBox{
								CheckState: walk.CheckChecked,
							},
							PushButton{
								AssignTo: &saveToLeft,
								Text:     "<-",
								OnClicked: func() {
									saveToPrev()
								},
							},
							LineEdit{
								AssignTo: &saveTo,
								Text:     saveToValues[0],
								Enabled:  false,
							},

							PushButton{
								AssignTo: &saveToRight,
								Text:     "->",
								OnClicked: func() {
									saveToNext()

								},
							},
						},
					},
					PushButton{
						AssignTo: &toggleGhostMod,
						Text:     "Start Ghost Mod",
						OnClicked: func() {
							if ghostMod {
								select {
								case stopGhostMod <- true:
								default:
								}
								toggleReplayMode.SetVisible(false)
								toggleReplayMode.SetText("Replay Mode [OFF]")
								toggleSpeedrunMode.SetVisible(false)
								toggleSpeedrunMode.SetText("Speedrun Mode [OFF]")
								ghostMod = false
								toggleGhostMod.SetText("Start Ghost Mod")
							} else {
								go ghostTool(stopGhostMod)
								ghostMod = true
								replayMode = false
								speedrunMode = false
								toggleSpeedrunMode.SetVisible(true)
								saveButton.SetVisible(false)
								toggleSpeedrunMode.SetText("Speedrun Mode [OFF]")
								toggleReplayMode.SetText("Replay Mode [OFF]")

								toggleReplayMode.SetVisible(true)
								toggleGhostMod.SetText("Stop Ghost Mod")

							}
						},
					},
					PushButton{
						AssignTo: &toggleReplayMode,
						Visible:  false,
						Text:     "Replay Mode [OFF]",
						OnClicked: func() {
							if replayMode {
								replayMode = false
								toggleReplayMode.SetText("Replay Mode [OFF]")
							} else {
								replayMode = true
								toggleReplayMode.SetText("Replay Mode [ON]")
							}
						},
					},
					HSplitter{
						Children: []Widget{
							PushButton{
								AssignTo: &toggleSpeedrunMode,
								Visible:  false,
								Text:     "Speedrun Mode [OFF]",
								OnClicked: func() {
									if speedrunMode {
										speedrunMode = false
										saveButton.SetVisible(false)
										toggleSpeedrunMode.SetText("Speedrun Mode [OFF]")

									} else {
										speedrunMode = true
										saveButton.SetVisible(true)
										toggleSpeedrunMode.SetText("Speedrun Mode [ON]")
										resetCurrentAttempt()

									}
								},
							},
							PushButton{
								AssignTo: &saveButton,
								Visible:  false,
								Text:     "Save Last Run",
								OnClicked: func() {
									popupWindow :=
										MainWindow{

											Title:   "Save Last Speedrun",
											MinSize: Size{300, 200},
											Layout:  VBox{},
											Children: []Widget{
												PushButton{
													Text: "Save",
													OnClicked: func() {
														saveCurrentAttempt(savePackTextEdit.Text())
													},
												},
												LineEdit{
													AssignTo: &savePackTextEdit,
													Text:     "",
													Enabled:  true,
												},
											},
										}
									popupWindow.Run()
								},
							},
						},
					},

					PushButton{
						Text: "Pack Manager",
						OnClicked: func() {
							openPackManager()
						},
					},
					PushButton{
						AssignTo: &toggleRewinder,
						Text:     "Rewinder [OFF]",
						OnClicked: func() {
							saveStates = false
							stateButton.SetText("Save Engine [Off]")
							if rewinderEnabled == false {
								go rewinder()
								toggleRewinder.SetText("Rewinder [ON]")
							} else {
								rewinderEnabled = false
								toggleRewinder.SetText("Rewinder [OFF]")
							}
						},
					},
				},
			}.Run()); err != nil {
			log.Fatal(err)
		}
		//go saveEngine()

	}

}

func openPackManager() {
	var usernameField *walk.LineEdit
	var passwordField *walk.LineEdit
	var packList *walk.TextEdit
	var downloadPackName *walk.LineEdit
	var loginComponents *walk.Splitter
	var manageComponents *walk.Splitter
	var uploadLine *walk.Splitter
	var toUpload = 0
	var toUploadLine *walk.LineEdit
	token := ""
	req_url := "http://www.spades.cloud"
	popupWindow :=
		MainWindow{
			Title:   "Pack Manager",
			MinSize: Size{300, 200},
			Layout:  VBox{},
			Children: []Widget{
				VSplitter{
					AssignTo: &loginComponents,
					Children: []Widget{
						HSplitter{
							Children: []Widget{
								Label{
									Text: "Username: ",
								},
								LineEdit{
									AssignTo: &usernameField,
								},
							},
						},
						HSplitter{
							Children: []Widget{
								Label{
									Text: "Password: ",
								},
								LineEdit{
									AssignTo:     &passwordField,
									PasswordMode: true,
								},
							},
						},
						HSplitter{
							Children: []Widget{
								PushButton{
									Text: "Login",
									OnClicked: func() {
										form := url.Values{
											"username": {usernameField.Text()},
											"password": {passwordField.Text()},
										}
										body := bytes.NewBufferString(form.Encode())
										rsp, err := http.Post(req_url+"/users/request/login", "application/x-www-form-urlencoded", body)
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}
										defer rsp.Body.Close()
										bodyByte, err := ioutil.ReadAll(rsp.Body)
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}
										var loginRes LoginResponse
										_ = json.Unmarshal(bodyByte, &loginRes)
										if loginRes.Login == "success" {
											token = loginRes.Token
											loginComponents.SetVisible(false)
											manageComponents.SetVisible(true)
											if len(saveToValues) == 1 {
												uploadLine.SetVisible(false)
											}
											form := url.Values{
												"token": {token},
											}
											body := bytes.NewBufferString(form.Encode())
											rsp, err := http.Post(req_url+"/packs/getPacks", "application/x-www-form-urlencoded", body)
											if err != nil {
												fmt.Printf("Err: %s\n", err)
											}

											bodyByte, err := ioutil.ReadAll(rsp.Body)

											var dat []map[string]string
											_ = json.Unmarshal(bodyByte, &dat)
											for i := 0; i < len(dat); i++ {
												packList.SetText(packList.Text() + dat[i]["pack"] + "\r\n")
											}

											defer rsp.Body.Close()
											//fmt.Printf("Token: %s\n", loginRes.Token)
										} else {
											w32.MessageBox(0, loginRes.Error, "Login", 0)
											//fmt.Printf("Err: %s\n", loginRes.Error)
										}

										//loginComponents.SetVisible(false)
									},
								},
								PushButton{
									Text: "Signup",
									OnClicked: func() {
										form := url.Values{
											"username": {usernameField.Text()},
											"password": {passwordField.Text()},
										}
										body := bytes.NewBufferString(form.Encode())
										rsp, err := http.Post(req_url+"/users/request/signup", "application/x-www-form-urlencoded", body)
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}
										defer rsp.Body.Close()
										bodyByte, err := ioutil.ReadAll(rsp.Body)
										w32.MessageBox(0, string(bodyByte), "Signup", 0)
									},
								},
							},
						},
					},
				},
				VSplitter{
					AssignTo: &manageComponents,
					Visible:  false,
					Children: []Widget{
						HSplitter{
							Children: []Widget{
								LineEdit{
									AssignTo: &downloadPackName,
									Text:     "",
								},
								PushButton{
									Text: "Download",
									OnClicked: func() {

										form := url.Values{
											"token": {token},
										}

										body := bytes.NewBufferString(form.Encode())

										rsp, err := http.Post(req_url+"/packs/getPacks/"+downloadPackName.Text(), "application/x-www-form-urlencoded", body)
										defer rsp.Body.Close()
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}

										bodyByte, err := ioutil.ReadAll(rsp.Body)

										if string(bodyByte) == "error" {
											w32.MessageBox(0, "Invalid Pack? Copy packname from list Below", "pack error", 0)
											return
										}

										insertStatement, _ := db.Prepare("REPLACE INTO packs (name, owner, times) VALUES(?, 0, ?)")
										insertStatement.Exec(downloadPackName.Text(), string(bodyByte))
										insertStatement.Close()

										countReq, _ := db.Query("select * from sqlite_master where type='table' and name='[" + downloadPackName.Text() + "]'")

										if countReq.Next() {

											statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS [" + downloadPackName.Text() + "] (level INTEGER, char INTEGER, data BLOB)")
											if err != nil {
												fmt.Printf("Error Creating Table: %s\n", err)
											}
											defer statement.Close()
											_, err = statement.Exec()
										} else {

											statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS [" + downloadPackName.Text() + "] (level INTEGER, char INTEGER, data BLOB)")
											if err != nil {
												fmt.Printf("Error Creating Table: %s\n", err)
											}
											defer statement.Close()
											_, err = statement.Exec()
											text := strings.Replace(downloadPackName.Text(), ".", "", -1)
											indexStatement, err := db.Prepare("CREATE UNIQUE INDEX if not exists idx_level_" + text + " ON [" + downloadPackName.Text() + "] (level)")
											if err != nil {
												fmt.Printf("Error Creating Index: %s\n", err)
											}
											indexStatement.Exec()
											defer indexStatement.Close()
										}
										countReq.Close()

										form = url.Values{
											"token": {token},
										}

										body = bytes.NewBufferString(form.Encode())

										rsp, err = http.Post(req_url+"/packs/getGhosts/"+downloadPackName.Text(), "application/x-www-form-urlencoded", body)
										defer rsp.Body.Close()
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}

										bodyByte, err = ioutil.ReadAll(rsp.Body)

										if string(bodyByte) == "error" {
											w32.MessageBox(0, "Invalid Pack? Copy packname from list Below", "pack error", 0)
											return
										}
										var data []map[string]string
										_ = json.Unmarshal(bodyByte, &data)
										fmt.Printf("Datalen: %d\n", len(data))
										for i := 0; i < len(data); i++ {

											level, _ := strconv.Atoi(data[i]["level"])
											char, _ := strconv.Atoi(data[i]["char"])

											insertGhost, err := db.Prepare("REPLACE INTO [" + downloadPackName.Text() + "] (level, char, data) VALUES(?, ?, ?)")
											if err != nil {
												fmt.Printf("Error preparing ghost data: %s\n", err)
											}
											defer insertGhost.Close()
											_, err = insertGhost.Exec(level, char, data[i]["data"])
											if err != nil {
												fmt.Printf("Error Inserting ghost data: %s\n", err)
											}
											fmt.Printf("Downloaded %f percent \n", (float64(i)*100)/float64(len(data)))
										}
										fmt.Printf("Downloaded 100 percent, DONE!\n")
										var exists = false
										for _, b := range raceAgainstValues {
											if b == downloadPackName.Text() {
												exists = true
											}
										}
										if !exists {
											raceAgainstValues = append(raceAgainstValues, downloadPackName.Text())
											raceAgainstSel = len(raceAgainstValues) - 1
											raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
										}

										/*insertStatement, _ := db.Prepare("REPLACE INTO packs (name, owner, times) VALUES(?, 0, ?)")
										defer insertStatement.Close()
										insertStatement.Exec(downloadPackName.Text(), string(bodyByte))*/

									},
								},
							},
						},

						TextEdit{
							AssignTo: &packList,
							Text:     "",
							ReadOnly: true,
						},
						HSplitter{
							AssignTo: &uploadLine,
							Children: []Widget{
								PushButton{
									Text: "<-",
									OnClicked: func() {
										if toUpload == 1 {
											toUpload = len(saveToValues) - 1

										} else {
											toUpload--
										}
										toUploadLine.SetText(saveToValues[toUpload])
									},
								},
								LineEdit{
									AssignTo: &toUploadLine,
									Text:     saveToValues[toUpload],
									Enabled:  false,
								},
								PushButton{
									Text: "->",
									OnClicked: func() {
										if toUpload == len(saveToValues)-1 {
											toUpload = 1

										} else {
											toUpload++
										}
										toUploadLine.SetText(saveToValues[toUpload])
									},
								},
								PushButton{
									Text: "Upload / Update",
									OnClicked: func() {
										getTimesStatement, _ := db.Prepare("SELECT times from packs where name = ?")
										rows, _ := getTimesStatement.Query(saveToValues[toUpload])
										defer getTimesStatement.Close()
										var stringTimes []byte
										for rows.Next() {
											rows.Scan(&stringTimes)

										}
										form := url.Values{
											"token":    {token},
											"packName": {saveToValues[toUpload]},
											"times":    {string(stringTimes)},
										}
										body := bytes.NewBufferString(form.Encode())
										_, err := http.Post(req_url+"/packs/upload/", "application/x-www-form-urlencoded", body)
										if err != nil {
											fmt.Printf("Err: %s\n", err)
										}

										getGhosts, err := db.Prepare("SELECT level, char, data from [" + saveToValues[toUpload] + "]")
										if err != nil {
											fmt.Printf("Error %s\n", err)
										}
										ghostRows, err := getGhosts.Query()
										if err != nil {
											fmt.Printf("Error %s\n", err)
										}
										defer getGhosts.Close()
										for ghostRows.Next() {
											var level int
											var char int
											var ghostData []byte
											ghostRows.Scan(&level, &char, &ghostData)
											form := url.Values{
												"token":    {token},
												"packName": {saveToValues[toUpload]},
												"level":    {strconv.Itoa(level)},
												"char":     {strconv.Itoa(char)},
												"data":     {string(ghostData)},
											}
											body := bytes.NewBufferString(form.Encode())
											_, err := http.Post(req_url+"/packs/UploadGhostData/", "application/x-www-form-urlencoded", body)
											if err != nil {
												fmt.Printf("Err: %s\n", err)
											}
											fmt.Printf("Uploaded Level: %d\n", level)
										}
										fmt.Printf("Done Uploading\n")
									},
								},
							},
						},
					},
				},
			},
		}
	popupWindow.Run()
}

var replayMode bool
var speedrunMode bool

func ghostTool(stopChan chan bool) {
	replayMode = false
	speedrunMode = false
	var replayHidden = false
	var lastMenu = -1
	lastLevelLoaded := -1
	/*
		0 = menus
		1 = loading level
		2 = loading level
		7 = finished loading (happens on restart)
		8 = beat level
		14 = game over?
		16 = game running
		17 = pause
		21 = exit

	*/

	var timeAsString = ""
	var replayToSave = make(map[string]GhostData)
	var replayToPlay = make(map[string]GhostData)
	var oldMili = -1
	var level = -1
	var char = -1
	var objOrigData1Loc uint32
	var objOrigData2Loc uint32
	var location uint32
	var longestAttempt = 0
	var attemptTime = 0
	var copyExists = false
	var copyHidden = false
	var lastLevelSaved = -1
	var lastAction = -1

	var levelLostTime = 0
	var speedrunTimeOffset = 0

LOOP:
	for {
		select {
		case stop := <-stopChan:
			if stop {
				fmt.Printf("Ending ghost Loop\n")
				break LOOP
			}
		default:
		}
		menu, _ := w32.ReadProcessMemory(handle, 0x01934BE0, 1)
		menuItem := menu[0]
		menuInt := int(menuItem)
		if menuInt != lastMenu {
			fmt.Printf("Menu: %d\n", menuInt)
			lastMenu = menuInt
			if menuInt == 16 {
				if len(replayToPlay) != 0 {
					if !replayMode {
						if !copyExists {
							if showGhost {
								if replayMode == false {
									simpleCopy(location, handle)
									copyExists = true
								}

							}
						} else if copyHidden {
							copyHidden = false
							addToGame()
						}
					} else {

						if replayHidden {
							addReplayToGame(location)
							replayHidden = false
						}

					}
				}

			}
			if menuInt == 8 {
				//saved stored data
				if replayMode == false {
					if !speedrunMode {
						if saveGhost {
							if currTimes.Times[level] == 0 || findTime(timeAsString) < currTimes.Times[level] {
								currTimes.Times[level] = findTime(timeAsString)
								timesToSave, _ := json.Marshal(currTimes)
								timesToSaveString := string(timesToSave)

								updateStatement, _ := db.Prepare("UPDATE packs SET times = ? WHERE name = ?")
								updateStatement.Exec(timesToSaveString, saveToValues[saveToSel])

								insertStatement, err := db.Prepare("REPLACE INTO [" + saveToValues[saveToSel] + "] (level, char, data) VALUES(?, ?, ?)")
								if err != nil {
									fmt.Printf("Error: %s\n", err)
								}
								data, err := json.Marshal(replayToSave)

								_, err = insertStatement.Exec(level, char, data)
								if err != nil {
									fmt.Printf("Error: %s\n", err)
								}
								fmt.Printf("new Best Time for this pack!\n")
							}
						}
					} else {
						if speedrunRace {
							if currTimes.Times[level] != 0 {
								speedrunTimeOffset = addTimes(addTimes(speedrunTimeOffset, findTime(timeAsString)), addTimes(levelLostTime, -1*raceTimes.Times[level]))
							}
							//speedrunTimeOffset += findTime(timeAsString) + levelLostTime - raceTimes.Times[level]
						}
						if saveGhost {
							if currTimes.Times[level] == 0 || addTimes(findTime(timeAsString), levelLostTime) < currTimes.Times[level] {
								currTimes.Times[level] = addTimes(findTime(timeAsString), levelLostTime)
								timesToSave, _ := json.Marshal(currTimes)
								timesToSaveString := string(timesToSave)

								updateStatement, _ := db.Prepare("UPDATE packs SET times = ? WHERE name = ?")
								updateStatement.Exec(timesToSaveString, saveToValues[saveToSel])

								insertStatement, err := db.Prepare("REPLACE INTO [" + saveToValues[saveToSel] + "] (level, char, data) VALUES(?, ?, ?)")
								if err != nil {
									fmt.Printf("Error: %s\n", err)
								}
								data, err := json.Marshal(replayToSave)

								_, err = insertStatement.Exec(level, char, data)
								if err != nil {
									fmt.Printf("Error: %s\n", err)
								}
								fmt.Printf("new Best Time for this pack!\n")
							}
						}
						speedrunTimes.Times[level] = addTimes(findTime(timeAsString), levelLostTime)
						timesToSave, _ := json.Marshal(speedrunTimes)
						timesToSaveString := string(timesToSave)

						updateStatement, _ := db.Prepare("UPDATE packs SET times = ? WHERE name = ?")
						updateStatement.Exec(timesToSaveString, "Current_Attempt")

						insertStatement, err := db.Prepare("REPLACE INTO [" + "Current_Attempt" + "] (level, char, data) VALUES(?, ?, ?)")
						if err != nil {
							fmt.Printf("Error: %s\n", err)
						}
						data, err := json.Marshal(replayToSave)

						_, err = insertStatement.Exec(level, char, data)
						if err != nil {
							fmt.Printf("Error: %s\n", err)
						}
					}
				}

			}
			if menuInt == 10 {

			}
			if menuInt == 0 {

			}
			if menuInt < 4 {
				replayToSave = make(map[string]GhostData)
			}
			if menuInt == 7 {
				location, _ = w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
				objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)
				objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)
				if replayMode == true {
					//w32.WriteProcessMemory(handle, location+0x10, []uint8{0, 0, 0, 0}, uint(0x4))
				}
				currChar, _ := w32.ReadProcessMemory(handle, 0x1934B80, uint(1))
				char = int(currChar[0])
				currLevel, _ := w32.ReadProcessMemory(handle, 0x01934B70, uint(1))
				level = int(currLevel[0])
				if level != lastLevelLoaded {
					levelLostTime = 0
					lastLevelLoaded = level
				} else {
					currTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, 3)
					min := int(currTime[0])
					sec := int(currTime[1])
					mili := int(currTime[2])
					minS := strconv.Itoa(min)

					if len(minS) == 1 {
						minS = "0" + minS
					}
					secS := strconv.Itoa(sec)
					if len(secS) == 1 {
						secS = "0" + secS
					}
					miliS := strconv.Itoa(mili)
					if len(miliS) == 1 {
						miliS = "0" + miliS
					}
					timeString := minS + secS + miliS

					lastTime := findTime(timeAsString)
					nowTime := findTime(timeString)

					timeDiff := addTimes(lastTime, -1*nowTime)
					fmt.Printf("Last  time: %d\n", lastTime)
					fmt.Printf("Now Time: %d\n", nowTime)
					if speedrunMode {
						levelLostTime = addTimes(levelLostTime, timeDiff)
					}

				}
				fmt.Printf("lost level time: %d\n", levelLostTime)
				fmt.Printf("lost level time string: %s\n", findTimeString(levelLostTime))
				if level == lastLevelSaved {
					if raceAgainstValues[raceAgainstSel] == "Last Attempt" {
						longestAttempt = 0
						for k, v := range replayToSave {
							replayToPlay[k] = v
						}

					} else if raceAgainstValues[raceAgainstSel] == "Longest Attempt" {
						if attemptTime > longestAttempt {
							for k, v := range replayToSave {
								replayToPlay[k] = v
							}
							longestAttempt = attemptTime
						}

					} else {

					}
				} else {
					if raceAgainstValues[raceAgainstSel] == "Last Attempt" || raceAgainstValues[raceAgainstSel] == "Longest Attempt" {
						replayToPlay = make(map[string]GhostData)
					} else {

					}
				}
				lastLevelSaved = level

			}
			if menuInt == 21 {
				if raceAgainstValues[raceAgainstSel] == "Last Attempt" {
					longestAttempt = 0
					for k, v := range replayToSave {
						replayToPlay[k] = v
					}

				} else if raceAgainstValues[raceAgainstSel] == "Longest Attempt" {
					if attemptTime > longestAttempt {
						for k, v := range replayToSave {
							replayToPlay[k] = v
						}
						longestAttempt = attemptTime
					}

				} else {

				}
			}
			if menuInt == 1 {
				currLevel, _ := w32.ReadProcessMemory(handle, 0x01934B70, uint(1))
				level = int(currLevel[0])

				if raceAgainstValues[raceAgainstSel] == "Last Attempt" || raceAgainstValues[raceAgainstSel] == "Longest Attempt" {

				} else {
					fmt.Printf("Start Loading\n")
					getStatement, err := db.Prepare("SELECT data from [" + raceAgainstValues[raceAgainstSel] + "] where level = ?")
					if err != nil {
						fmt.Printf("Error loading level: %s\n", err)
					}
					rows, err := getStatement.Query(level)
					if err != nil {
						fmt.Printf("Error: %s\n", err)
					}
					fmt.Printf("Level: %d, pack: %s\n", level, raceAgainstValues[raceAgainstSel])
					replayToPlay = make(map[string]GhostData)

					for rows.Next() {
						fmt.Printf("Start Loading\n")
						var tempData []byte
						rows.Scan(&tempData)
						_ = json.Unmarshal(tempData, &replayToPlay)
						fmt.Printf("Loaded len: %d\n", len(replayToPlay))
					}
				}

			}
			if menuInt == 17 || menuInt == 14 || menuInt == 8 || menuInt == 5 || menuInt == 9 {

				if copyExists {
					removeFromGame()
					copyHidden = true
				}
				if replayMode {
					if replayHidden == false {
						replayHidden = true
						removeReplayFromGame(location)
					}
				}
			}

			if menuInt < 4 {
				copyHidden = false
				copyExists = false
				replayHidden = false
			}

		}
		if menuInt == 17 || menuInt == 9 {
			//currTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, 3)
			//timeAsString = getCurrTimeString(currTime)
		}
		if speedrunMode {
			if menuInt == 0 {
				lastLevelLoaded = -1
				inMenu, _ := w32.ReadProcessMemory(handle, 0x01D7BB14, uint(1))
				inMenuInt := int(inMenu[0])
				if inMenuInt == 12 {
					speedrunTimeOffset = 0
					levelLostTime = 0
					resetCurrentAttempt()
					time.Sleep(time.Second)
				}
			}
		} else {
			if menuInt == 1 || menuInt == 0 {

				speedrunTimeOffset = 0
				levelLostTime = 0
			}

		}
		if menuInt == 16 {
			currTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, 3)
			timeAsString = getCurrTimeString(currTime)
			mili := int(currTime[2])
			min := int(currTime[0])
			sec := int(currTime[1])
			attemptTime = findTime(timeAsString)
			var offsetTimeAsString string
			storeTime := findTimeString(addTimes(attemptTime, levelLostTime))
			if speedrunMode {

				offsetTimeAsString = findTimeString(addTimes(addTimes(attemptTime, levelLostTime), speedrunTimeOffset))
			} else {
				offsetTimeAsString = timeAsString
			}
			action, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, uint(1))
			actionInt := int(action[0])
			if actionInt != lastAction {

				lastAction = actionInt
			}
			if copyExists {
				if actionInt >= 19 && actionInt < 30 {

					if copyHidden == false {
						removeFromGame()
						copyHidden = true
					}
				} else {

					if copyHidden {
						addToGame()
						copyHidden = false
					}
				}
			}
			if mili != oldMili {
				oldMili = mili
				if replayMode == false {
					tempData := new(GhostData)
					data, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, 0x30)

					tempData.Obj1 = data
					data, _ = w32.ReadProcessMemory(handle, objOrigData2Loc+0x174, Obj2Size)
					tempData.Obj2 = data

					replayToSave[storeTime] = *tempData

				}

				if showGhost {
					if val, ok := replayToPlay[offsetTimeAsString]; ok {
						if replayMode {

							/*  THESE SETTINGS ARE WORKING FOR CG AND MH */
							if sec == 0 && mili == 0 && min == 0 {

							} else if sec == 0 && mili == 1 && min == 0 {
								//w32.WriteProcessMemory(handle, location+0x10, []uint8{0, 0, 0, 0}, uint(0x4))
							} else {

								w32.WriteProcessMemory(handle, objOrigData1Loc+0x0, val.Obj1[0x0:0x20], 0x20-0x0)
								w32.WriteProcessMemory(handle, objOrigData2Loc+0x174, val.Obj2, Obj2Size)

							}

						} else {
							if char == 4 || char == 5 {

								w32.WriteProcessMemory(handle, obj1Loc+uint32(0x14), val.Obj1[0x14:0x20], 0x20-0x14)
							} else {

								copy(val.Obj1[0x0:0x8], []uint8{0, 0, 0, 0, 0, 0, 0, 0})
								w32.WriteProcessMemory(handle, obj1Loc, val.Obj1, 0x30)
								w32.WriteProcessMemory(handle, obj2Loc+0x174, val.Obj2, Obj2Size)
							}
						}

					}

				}
			}
		}

		time.Sleep(1000000)
	}

}
func getCurrTimeString(time []byte) string {
	min := int(time[0])
	sec := int(time[1])
	mili := int(time[2])
	minS := strconv.Itoa(min)

	if len(minS) == 1 {
		minS = "0" + minS
	}
	secS := strconv.Itoa(sec)
	if len(secS) == 1 {
		secS = "0" + secS
	}
	miliS := strconv.Itoa(mili)
	if len(miliS) == 1 {
		miliS = "0" + miliS
	}
	return minS + secS + miliS
}
func resetCurrentAttempt() {
	deleteStatement, _ := db.Prepare("DELETE FROM packs where name = ?")
	defer deleteStatement.Close()
	deleteStatement.Exec("Current_Attempt")

	deleteLevels, _ := db.Prepare("DROP TABLE [Current_Attempt]")
	defer deleteLevels.Close()
	deleteLevels.Exec()
	raceSelected := raceAgainstSel
	packExists := false
	for i := 0; i < len(raceAgainstValues); i++ {
		if raceAgainstValues[i] == "Current_Attempt" {
			packExists = true
		}
	}
	createGhostPack("Current_Attempt")
	if packExists {
		raceAgainstValues = raceAgainstValues[:len(raceAgainstValues)-1]
	}
	raceAgainstSel = raceSelected
	raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
	checkExistStatement, _ := db.Prepare("SELECT times from packs where name = ?")
	rows, _ := checkExistStatement.Query("Current_Attempt")
	defer rows.Close()
	for rows.Next() {
		var stringTimes []byte
		rows.Scan(&stringTimes)
		_ = json.Unmarshal(stringTimes, &speedrunTimes)
	}
	loadRacePack()

}
func loadRacePack() {
	fmt.Printf("Loading Pack\n")
	checkExistStatement, _ := db.Prepare("SELECT times from packs where name = ?")
	rows, _ := checkExistStatement.Query(raceAgainstValues[raceAgainstSel])
	defer rows.Close()
	speedrunRace = false
	for rows.Next() {
		speedrunRace = true
		var stringTimes []byte
		rows.Scan(&stringTimes)
		_ = json.Unmarshal(stringTimes, &raceTimes)
		fmt.Printf("Loaded Pack: %s\n", raceAgainstValues[raceAgainstSel])
		for i := 0; i < len(raceTimes.Times); i++ {
			fmt.Printf("%d: %d\n", i, raceTimes.Times[i])
		}
	}

}
func saveCurrentAttempt(name string) {
	isValidPackName := regexp.MustCompile(`^[A-Za-z1-9-_]+$`).MatchString
	if !isValidPackName(name) {
		w32.MessageBox(0, "Invalid Pack Name [A-B,a-b,1-9,_,-] allowed", "Error", 0)
		return
	}
	if name == "" {
		w32.MessageBox(0, "Invalid Pack Name [A-B,a-b,1-9,_,-] allowed", "Error", 0)
		return
	}
	checkExistStatement, _ := db.Prepare("SELECT name from packs WHERE name = ?")
	defer checkExistStatement.Close()
	rows, _ := checkExistStatement.Query(name)
	defer rows.Close()
	if rows.Next() {
		w32.MessageBox(0, "Pack Already Exists", "Error", 0)
	} else {
		renameStatement, err := db.Prepare("ALTER TABLE [Current_Attempt] RENAME TO [" + name + "]")
		if err != nil {
			fmt.Printf("Error Renaming: %s\n", err)
		}
		defer renameStatement.Close()
		renameStatement.Exec()
		updateStatement, err := db.Prepare("UPDATE packs SET name = ? WHERE name = 'Current_Attempt'")
		if err != nil {
			fmt.Printf("Error copying: %s\n", err)
		}
		defer updateStatement.Close()
		_, err = updateStatement.Exec(name)
		if err != nil {
			fmt.Printf("Error execing: %s\n", err)
		}
		for i := 0; i < len(raceAgainstValues); i++ {
			if raceAgainstValues[i] == "Current_Attempt" {
				raceAgainstValues[i] = name
			}
		}
		createGhostPack("Current_Attempt")
	}
}

func gui() {

}
func findTime(time string) int {
	min, _ := strconv.Atoi(time[:2])
	sec, _ := strconv.Atoi(time[2:4])
	mil, _ := strconv.Atoi(time[4:])
	totalTime := min*60*100 + sec*100 + mil
	return totalTime
}
func findTimeString(time int) string {
	mins := int(time / 6000)

	time -= mins * 60 * 100
	secs := int(time / 100)

	time -= secs * 100
	milis := time

	var minString string
	var secString string
	var miliString string
	if mins < 10 {
		minString = "0" + strconv.Itoa(mins)
	} else {
		minString = strconv.Itoa(mins)
	}
	if secs < 10 {
		secString = "0" + strconv.Itoa(secs)
	} else {
		secString = strconv.Itoa(secs)
	}
	if milis < 10 {
		miliString = "0" + strconv.Itoa(milis)
	} else {
		miliString = strconv.Itoa(milis)
	}
	return minString + secString + miliString
}
func addTimes(time1 int, time2 int) int {
	time1Mins := int(time1 / 6000)
	time1 -= time1Mins * 60 * 100
	time1Secs := int(time1 / 100)
	time1 -= time1Secs * 100
	time1Milis := time1

	time2Mins := int(time2 / 6000)
	time2 -= time2Mins * 60 * 100
	time2Secs := int(time2 / 100)
	time2 -= time2Secs * 100
	time2Milis := time2

	endMilis := time1Milis + time2Milis
	endSecs := 0
	endMins := 0

	if endMilis >= 60 {
		endMilis -= 60
		endSecs++
	}
	endSecs += time1Secs + time2Secs
	if endSecs >= 60 {
		endSecs -= 60
		endMins++
	}
	endMins += time1Mins + time2Mins
	endTime := endMins*60*100 + endSecs*100 + endMilis
	if int(math.Abs(float64(endTime)))%100 >= 60 {
		if endTime > 0 {
			endTime += 100
			endTime -= 60
		} else {
			endTime -= 100
			endTime += 60
		}

	}
	return endTime
}
func loadSavePack() {
	packName := saveToValues[saveToSel]
	checkExistStatement, _ := db.Prepare("SELECT times from packs where name = ?")
	rows, _ := checkExistStatement.Query(packName)
	loadedPack := false
	for rows.Next() {
		loadedPack = true
		var stringTimes []byte
		rows.Scan(&stringTimes)
		_ = json.Unmarshal(stringTimes, &currTimes)
	}
	if !loadedPack {
		return
	}

}
func loadPacks() {
	checkExistStatement, _ := db.Prepare("SELECT name,owner from packs")
	defer checkExistStatement.Close()
	rows, _ := checkExistStatement.Query()

	for rows.Next() {
		var packName string
		var owner int
		rows.Scan(&packName, &owner)
		if owner == 1 {
			saveToValues = append(saveToValues, packName)
		}
		raceAgainstValues = append(raceAgainstValues, packName)
	}
}
func createGhostPack(name string) {
	fmt.Printf("Creating %s\n", name)
	isValidPackName := regexp.MustCompile(`^[A-Za-z1-9-_]+$`).MatchString
	if !isValidPackName(name) {
		w32.MessageBox(0, "Invalid Pack Name [A-B,a-b,1-9,_,-] allowed", "Error", 0)
		return
	}
	if name == "" {
		w32.MessageBox(0, "Invalid Pack Name [A-B,a-b,1-9,_,-] allowed", "Error", 0)
		return
	}
	checkExistStatement, _ := db.Prepare("SELECT name from packs WHERE name = ?")
	defer checkExistStatement.Close()
	rows, _ := checkExistStatement.Query(name)
	defer rows.Close()
	if rows.Next() {
		w32.MessageBox(0, "Pack Already Exists", "Error", 0)
	} else {
		levelCount := 70
		blankData := make([]int, levelCount)
		blankTimes := Result{blankData}
		blankTimesData, _ := json.Marshal(blankTimes)
		blankTimesString := string(blankTimesData)
		insertStatement, _ := db.Prepare("INSERT INTO packs (name, owner, times) VALUES(?, 1, ?)")
		defer insertStatement.Close()
		insertStatement.Exec(name, blankTimesString)
		statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS [" + name + "] (level INTEGER, char INTEGER, data BLOB)")
		defer statement.Close()
		statement.Exec()
		indexStatement, _ := db.Prepare("CREATE UNIQUE INDEX IF NOT EXISTS idx_level_" + name + " ON [" + name + "] (level)")
		indexStatement.Exec()
		defer indexStatement.Exec()

		if raceAgainst != nil {
			raceAgainstValues = append(raceAgainstValues, name)
			raceAgainstSel = len(raceAgainstValues) - 1
			raceAgainst.SetText(raceAgainstValues[raceAgainstSel])

		}
		saveToValues = append(saveToValues, name)

	}
}
func checkCurrentPack() {

}

func raceAgainstPrev() {
	if raceAgainstSel == 0 {
		raceAgainstSel = len(raceAgainstValues) - 1
		raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
	} else {
		raceAgainstSel--
		raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
	}
	if speedrunMode {
		loadRacePack()
	}
}

func raceAgainstNext() {

	if raceAgainstSel < len(raceAgainstValues)-1 {
		raceAgainstSel++
		raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
	} else {
		raceAgainstSel = 0
		raceAgainst.SetText(raceAgainstValues[raceAgainstSel])
	}
	if speedrunMode {
		loadRacePack()
	}
}

func saveToPrev() {
	if saveToSel == 0 {
		saveToSel = len(saveToValues) - 1
		saveTo.SetText(saveToValues[saveToSel])

	} else {
		saveToSel--
		saveTo.SetText(saveToValues[saveToSel])
	}

	if saveToValues[saveToSel] == "Current_Attempt" {
		saveToPrev()
		return
	}
	if saveToValues[saveToSel] == "" {
		saveGhost = false
	} else {
		saveGhost = true
	}
	loadSavePack()
}

func saveToNext() {
	if saveToSel < len(saveToValues)-1 {
		saveToSel++
		saveTo.SetText(saveToValues[saveToSel])
	} else {
		saveToSel = 0
		saveTo.SetText(saveToValues[saveToSel])
	}
	if saveToValues[saveToSel] == "Current_Attempt" {
		saveToNext()
		return
	}
	if saveToValues[saveToSel] == "" {
		saveGhost = false
	} else {
		saveGhost = true
	}
	loadSavePack()
}

func openInstructions() {
	var msg = "Coming soon LUL\n"
	w32.MessageBox(0, msg, "SPaDeS-Tools", 0)
}
func removeFromGame() {
	emptyData := make([]uint8, 0x50-0x8)
	for i := 0; i < len(emptyData); i++ {
		emptyData[i] = 0
	}
	for i := 0; i < len(playersInGame); i++ {
		tempData, _ := w32.ReadProcessMemory(handle, insertLoc+uint32(playersInGame[i].playerID*0x1000)+0x8, uint(len(emptyData)))
		playersInGame[i].pointerData = tempData
		w32.WriteProcessMemory(handle, insertLoc+uint32(playersInGame[i].playerID*0x1000)+0x8, emptyData, uint(len(emptyData)))
	}

}
func removeReplayFromGame(location uint32) {
	oldPointerData, _ = w32.ReadProcessMemory(handle, location, 0x50)
	emptyData := make([]uint8, 0x50-0x8)
	for i := 0; i < len(emptyData); i++ {
		emptyData[i] = 0
	}
	w32.WriteProcessMemory(handle, location+0x8, emptyData, uint(len(emptyData)))

}

func addToGame() {
	for i := 0; i < len(playersInGame); i++ {
		w32.WriteProcessMemory(handle, insertLoc+uint32(playersInGame[i].playerID*0x1000)+0x8, playersInGame[i].pointerData, uint(0x50-0x8))
	}
}
func addReplayToGame(location uint32) {
	w32.WriteProcessMemory(handle, location+0x8, oldPointerData[0x8:], uint(0x50-0x8))
}

var objOrigData1Loc uint32

func raceOutputter(menuChan chan int, levelChan chan int, existsChan chan bool, conn net.Conn, isClient bool, exitedLevelChan chan bool, closeChan chan bool) {
	var lastMenu = -1

	var currLevel = -1
	var currChar = -1
	var validLevel = false

	var objOrigData2Loc uint32
	var location uint32
	var otherExists = false
	var printOnce = true
LOOP:
	for {
		select {
		case exists := <-existsChan:
			otherExists = exists

		default:
		}

		menu, _ := w32.ReadProcessMemory(handle, 0x01934BE0, 1)
		menuItem := menu[0]
		menuInt := int(menuItem)
		if menuInt != lastMenu {

			select {
			case menuChan <- menuInt:
			default:
			}

			lastMenu = menuInt
			fmt.Printf("Menu: %d\n", menuInt)
			if menuInt == 16 {
				if otherExists {
					addToGame()

				}
				location, _ = w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
				objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)
				objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)
				tempLevel, _ := w32.ReadProcessMemory(handle, 0x01934b70, uint(1))
				tempChar, _ := w32.ReadProcessMemory(handle, 0x01934B80, uint(1))
				currChar = int(tempChar[0])

				currLevel = int(tempLevel[0])
				validLevel = validCloneLevel(currLevel, currChar)
				select {
				case levelChan <- currLevel:
				default:
				}
			}
			if menuInt == 17 {
				if otherExists {
					removeFromGame()
				}
			}
			if menuInt == 14 {
				if otherExists {
					removeFromGame()

				}
			}
			if menuInt == 8 {
				if otherExists {
					removeFromGame()
				}
			}
			if menuInt == 7 {
				location, _ = w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
				objOrigData1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x34)
				objOrigData2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, location+0x40)
			}
			if menuInt == 5 || menuInt == 9 {
				if otherExists {
					removeFromGame()
					otherExists = false
					select {
					case exitedLevelChan <- false:

					default:
					}
				}
			}

			if menuInt < 4 {
				validLevel = false
				otherExists = false
				select {
				case exitedLevelChan <- false:
				default:
				}
			}
		}
		select {
		case _ = <-closeChan:
			fmt.Printf("Breaking Sender\n")
			if otherExists {
				if menuInt == 16 || menuInt == 7 {
					removeFromGame()
				}
			}
			break LOOP
		default:
		}

		if lastMenu == 16 {
			if validLevel {
				var data0 = []byte{byte(currLevel)}
				data1, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, 0x30)
				data2, _ := w32.ReadProcessMemory(handle, objOrigData2Loc+0x174, Obj2Size)
				copy(data1[0x0:0x8], []uint8{0, 0, 0, 0, 0, 0, 0, 0})
				tempData := append(data0, data1...)
				data := append(tempData, data2...)
				w := bufio.NewWriter(conn)
				w.Write(data)
				w.Flush()
				if printOnce {

					if !isClient {

						printOnce = false
					}
				}
			}

			if otherExists {

			}
		}

		time.Sleep(1 * time.Second / 120)
	}
}

var playersInGame []Racer

func firstInstanceOfPlayer(player int) bool {
	for i := 0; i < len(playersInGame); i++ {
		if playersInGame[i].playerID == player {
			return false
		}
	}
	return true
}
func injector(data1 []byte, data2 []byte, newDataChan chan bool, closeChan chan bool, playerChan chan int) {
	var printOnce = true

	//var newData = false
LOOP1:
	for {
		select {
		case _ = <-closeChan:
			fmt.Printf("Breaking injector\n")
			break LOOP1
		default:
		}

		//min := int(currTime[0])

		select {
		case _ = <-newDataChan:
			select {
			case player := <-playerChan:
				menu, _ := w32.ReadProcessMemory(handle, 0x01934BE0, 1)
				menuItem := menu[0]
				menuInt := int(menuItem)
				if menuInt == 16 {

					tempChar, _ := w32.ReadProcessMemory(handle, 0x01934B80, uint(1))
					currChar := int(tempChar[0])
					if dataWriteLock == false {
						dataWriteLock = true
						if printOnce {

						}
						if currChar == 4 || currChar == 5 {
							w32.WriteProcessMemory(handle, obj1Loc+uint32(player*0x1000)+uint32(0x14), []byte(data1[0x14:0x20]), 0x20-0x14)
						} else {
							w32.WriteProcessMemory(handle, obj1Loc+uint32(player*0x1000), []byte(data1), 0x30)
							w32.WriteProcessMemory(handle, obj2Loc+uint32(player*0x1000)+0x174, []byte(data2), Obj2Size)
						}

						dataWriteLock = false
					}
				}
			default:
			}

		default:
		}

		time.Sleep(1000000)
	}

}
func race2(conn net.Conn, isClient bool) {
	playersInGame = make([]Racer, 0)
	defer conn.Close()
	fmt.Printf("Connection Made\n")
	exitedLevel := make(chan bool, 1)
	menu := make(chan int, 2)
	level := make(chan int, 1)
	exists := make(chan bool, 1)
	newData := make(chan bool, 1)
	closeInjector := make(chan bool, 1)
	closeSender := make(chan bool, 1)
	playerChan := make(chan int, 1)
	var currMenu = -1
	var currLevel = -1
	var otherLevel int = -2
	var othersExist = false
	var otherHidden = false

	var printOnce = true

	var d1 = make([]byte, 48)
	var d2 = make([]byte, Obj2Size)
	//var printOnce = false
	go raceOutputter(menu, level, exists, conn, isClient, exitedLevel, closeSender)
	go injector(d1, d2, newData, closeInjector, playerChan)
	r := bufio.NewReader(conn)

LOOP:
	for {

		fullBuf := make([]byte, 4+1+48+Obj2Size)
		readLen, err := r.Read(fullBuf)
		if err != nil {
			fmt.Printf("Err: %s\n", err)
			fmt.Printf("Connection Closed\n")
			if !isClient {

			} else {
				fmt.Printf("Setting connect text\n")
				connectedToServer = false
				connectButton.SetText("Connect to Server")
			}
			select {
			case closeInjector <- true:
			default:
			}
			select {
			case closeSender <- true:
			default:
			}
			break LOOP

		}

		if readLen != int(4+1+0x30+Obj2Size) {
			fmt.Printf("Wrong line length\n")
			continue
		}
		player := binary.LittleEndian.Uint32(fullBuf[0:4])
		lev := int(fullBuf[4])

		if printOnce {

			printOnce = false
			if isClient == true {

			}
		}

		if lev != 0 && lev < 91 {
			otherLevel = lev
		}
		if dataWriteLock == false {
			dataWriteLock = true
			copy(d1, fullBuf[4+1:48+4+1])

			copy(d2, fullBuf[4+1+48:])
			dataWriteLock = false

		} else {
			continue
		}
		menu, _ := w32.ReadProcessMemory(handle, 0x01934BE0, 1)
		menuItem := menu[0]
		menuInt := int(menuItem)
		if menuInt != currMenu {
			currMenu = menuInt
			if currMenu == 16 {
				time.Sleep(100000000)
				tempLevel, _ := w32.ReadProcessMemory(handle, 0x01934b70, uint(1))
				currLevel = int(tempLevel[0])

			}
		}
		select {
		case _ = <-exitedLevel:
			fmt.Printf("Setting sene to false exitedLevel channel \n")
			othersExist = false
		default:
		}

		select {
		case newLevel := <-level:
			currLevel = newLevel

		default:
		}

		if currMenu == 16 {

			if currLevel == otherLevel {

				if firstInstanceOfPlayer(int(player)) {
					tempPlayer := Racer{int(player), make([]byte, 0)}
					playersInGame = append(playersInGame, tempPlayer)
					othersExist = true

					simpleCopyToOffset(int(player))

					select {
					case exists <- true:

					default:
					}

					select {
					case playerChan <- int(player):
					default:
					}
					select {
					case newData <- true:
					default:
					}

				} else {
					action, _ := w32.ReadProcessMemory(handle, objOrigData1Loc, uint(1))
					actionInt := int(action[0])

					if othersExist {
						if actionInt >= 19 && actionInt < 30 {
							if otherHidden == false {
								removeFromGame()
								select {
								case exists <- false:

								default:
								}
								otherHidden = true
							}
						} else {

							if otherHidden {
								addToGame()
								select {
								case exists <- true:
								default:
								}
								otherHidden = false
							}
						}
					}
					select {
					case exists <- true:
					default:
					}
					select {
					case playerChan <- int(player):
					default:
					}
					select {
					case newData <- true:
					default:
					}

				}
			} else {
				if othersExist == true {
					select {
					case exists <- false:

					default:
					}
					othersExist = false
				}

			}

		} else if currMenu == 7 {

		} else if currMenu > 17 {

			if othersExist == true {
				select {
				case exists <- false:

				default:
				}
				othersExist = false
			}

		} else if currMenu < 4 {
			playersInGame = make([]Racer, 0)
			select {
			case exists <- false:

			default:
			}
			othersExist = false
		}

	}
}

var serverRunning = false
var ln net.Listener
var conn net.Conn
var connectedToServer = false

func connect(room string) {
	if room == "" {
		return
	}
	var err error
	conn, err = net.Dial("tcp", "174.5.140.220"+":8081")
	if err != nil {
		w32.MessageBox(0, "Error Connecting", "SPaDeS-Tools", 0)
		return
	}
	connectedToServer = true
	w := bufio.NewWriter(conn)
	w.Write([]byte(room + "\n"))
	w.Flush()
	race2(conn, true)
}
func startServer() {
	if serverRunning == false {
		serverRunning = true
		fmt.Println("Launching server...")
		ln, _ = net.Listen("tcp", ":8081")
		serverButton.SetText("Waiting for Connection...")
		for {
			var err error
			conn, err = ln.Accept()
			if err != nil {
				break
			}
			serverButton.SetText("Connected... (Disconnect)")
			time.Sleep(2 * time.Second)
			race2(conn, false)
		}
	} else {
		serverRunning = false
		ln.Close()
		conn.SetDeadline(time.Now())
		serverButton.SetText("Start Server")

	}
}

func connServer(server string) {
	if !connectedToServer {
		var err error
		conn, err = net.Dial("tcp", server+":8081")
		if err != nil {
			w32.MessageBox(0, "Error Connecting\nIs their port 8081 open for TCP?", "SPaDeS-Tools", 0)
			return
		}
		connectedToServer = true
		time.Sleep(2 * time.Second)
		connectButton.SetText("Connected... (Disconnect)")
		race2(conn, true)

	} else {
		conn.SetDeadline(time.Now())
	}
}

func checkByte(b byte, spot int) byte {
	spot = int(math.Pow(2, float64(spot)))
	spotUint := uint8(spot)
	spotByte := byte(spotUint)
	if b&spotByte > 0 {
		return 1
	} else {
		return 0
	}

}

var saveStates = false
var currentState = 0
var states []State

func saveState() {

	playerAddress, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
	tempState := State{}
	tempState.cam2 = getCamera2(handle)
	tempState.x = getX(playerAddress, handle)
	tempState.y = getY(playerAddress, handle)
	tempState.z = getZ(playerAddress, handle)
	tempState.horSpeed = getHorSpeed(playerAddress, handle)
	tempState.vertSpeed = getVertSpeed(playerAddress, handle)
	tempState.xRot = getXRot(playerAddress, handle)
	tempState.yRot = getYRot(playerAddress, handle)
	tempState.zRot = getZRot(playerAddress, handle)
	tempState.status = getStatus(playerAddress, handle)
	tempState.hangtime = getHangTime(playerAddress, handle)
	tempState.physics = getPhysics(playerAddress, handle)
	tempState.animFrames = getAnimFrames(playerAddress, handle)
	tempState.currAnim = getCurrAnim(playerAddress, handle)
	tempState.action = getAction(playerAddress, handle)
	tempState.hover = getHover(handle)
	tempState.momentum = getMomentum(handle)
	tempState.gametime = getTime(handle)
	tempState.gravity = getGravity(handle)
	tempLevel, _ := w32.ReadProcessMemory(handle, 0x01934b70, uint(1))
	tempState.level = int(tempLevel[0])

	tempState.ready = true
	states[currentState] = tempState
}
func loadState() {
	if !states[currentState].ready {
		return
	}
	tempLevel, _ := w32.ReadProcessMemory(handle, 0x01934b70, uint(1))
	if states[currentState].level != int(tempLevel[0]) {
		return
	}
	tempState := states[currentState]
	playerAddress, _ := w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
	for i := 0; i < 2; i++ {
		setXValue(playerAddress, handle, tempState.x)
		setYValue(playerAddress, handle, tempState.y)
		setZValue(playerAddress, handle, tempState.z)
		setHorSpeed(playerAddress, handle, tempState.horSpeed)
		setVertSpeed(playerAddress, handle, tempState.vertSpeed)
		setXRot(playerAddress, handle, tempState.xRot)
		setYRot(playerAddress, handle, tempState.yRot)
		setZRot(playerAddress, handle, tempState.zRot)
		setStatus(playerAddress, handle, tempState.status)
		setHangTime(playerAddress, handle, tempState.hangtime)
		setPhysics(playerAddress, handle, tempState.physics)
		setAnimFrames(playerAddress, handle, tempState.animFrames)
		setCurrAnim(playerAddress, handle, tempState.currAnim)
		setAction(playerAddress, handle, tempState.action)
		setHover(handle, tempState.hover)
		setMomentum(handle, tempState.momentum)
		setTime(handle, tempState.gametime)
		setCamera2(handle, tempState.cam2)
		setGravity(handle, tempState.gravity)
		time.Sleep(100 * time.Millisecond)
	}
}

func saveEngine() {
	saveStates = true

	var menu = -1
	//var controllerState xinput.State
	//go keyboardThread()
	xinput.Load()
	go inputHandler(&menu)
	for {
		if !saveStates {
			return
		}
		newMenuByte, _ := w32.ReadProcessMemory(handle, 0x01934BE0, uint(1))
		newMenu := int(newMenuByte[0])
		if newMenu != menu {
			menu = newMenu
		}
		time.Sleep(1000000)
		/*
			xinput.GetState(1, &controllerState)
			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(controllerState.Gamepad.Buttons))
		*/

	}
}

var oldMainPointer uint32
var loc uint32
var replayData []PlayerData
var recording = true
var currSpot = 0
var d1Loc uint32
var d2Loc uint32
var d2RecordOffset = 0x0
var d2RecordSize = 0x300
var rewinderEnabled = false
var jumpDistance = 5
var lastOffsetUsed = 0
var distanceFromPause = 0
var backDraw = 0
var lastDrawn = 0
var maxClones = 10

func buttonChange(button int, pressed bool, menu int) {
	if rewinderEnabled {
		tempChar, _ := w32.ReadProcessMemory(handle, 0x01934B80, uint(1))
		char := int(tempChar[0])
		if char == 4 || char == 5 {
			//return
		}

		if button == 6 {
			if pressed == true {
				if menu == 16 {
					if recording == false {
						if jumpDistance > 1 {
							jumpDistance--
							w32.WriteProcessMemoryAsUint32(handle, 0x0174B050, uint32(jumpDistance))
						}
					}
				}
			}
		}
		if button == 5 {
			if pressed == true {
				if menu == 16 {
					if recording == false {
						if jumpDistance < 100 {
							jumpDistance++
							w32.WriteProcessMemoryAsUint32(handle, 0x0174B050, uint32(jumpDistance))
						}
					}
				}
			}
		}
		if button == 0 {
			if pressed == true {
				if menu == 16 {

					tempOldPointer, _ := w32.ReadProcessMemoryAsUint32(handle, loc+0x10)
					fmt.Printf("Setting Pointer: %d\n", tempOldPointer)
					if tempOldPointer != 0 {
						oldMainPointer = tempOldPointer
					}
					w32.WriteProcessMemoryAsUint32(handle, loc+0x10, uint32(0))
					if recording == true {
						recording = false
						distanceFromPause = 0
						w32.WriteProcessMemoryAsUint32(handle, 0x0174B050, uint32(jumpDistance))
						//w32.WriteProcessMemory(handle, 0x01DE4664, []byte{255}, uint(1))
						w32.WriteProcessMemory(handle, 0x0174AFF7, []byte{1}, uint(1))
						var drawn = 0
						for i := ((currSpot - 1) / 10); i >= 0; i-- {

							if drawn >= 10 {

								break
							}
							simpleCopyToOffset(drawn)
							w32.WriteProcessMemory(handle, obj1Loc+uint32(drawn*0x1000), replayData[i*10].obj1, 0x30)
							w32.WriteProcessMemory(handle, obj2Loc+uint32(drawn*0x1000), replayData[i*10].obj2, uint(d2RecordSize))
							lastOffsetUsed = i
							lastDrawn = i * 10
							drawn++
						}
					} else {
						if currSpot > jumpDistance-1 {
							distanceFromPause += jumpDistance
							currSpot -= jumpDistance
							for {
								if distanceFromPause > 50+backDraw*10 {
									var spotToUse = lastDrawn - 10
									if spotToUse < 10 {
										break
									} else {
										w32.WriteProcessMemory(handle, obj1Loc+uint32((backDraw%10)*0x1000), replayData[spotToUse].obj1, 0x30)
										w32.WriteProcessMemory(handle, obj2Loc+uint32((backDraw%10)*0x1000), replayData[spotToUse].obj2, uint(d2RecordSize))
										lastDrawn = spotToUse
										backDraw++
									}
								} else {
									break
								}
							}
							w32.WriteProcessMemory(handle, d1Loc, replayData[currSpot].obj1, 0x30)
							w32.WriteProcessMemory(handle, d2Loc+uint32(d2RecordOffset), replayData[currSpot].obj2, uint(d2RecordSize))

						} else {
							fmt.Printf("Were at the start of recording\n")
						}
						//step back
					}
				}
			}
		}
		if button == 1 {
			if pressed == true {
				if menu == 16 {
					//step forward
					if recording == false {
						if currSpot < len(replayData)-jumpDistance {
							distanceFromPause -= jumpDistance
							currSpot += jumpDistance
							for {
								if distanceFromPause < 50+(backDraw-1)*10 {
									fmt.Printf("moving farthest clone\n")
									var spotToUse = lastDrawn + 100
									if spotToUse > currSpot+distanceFromPause-1 {
										fmt.Printf("Next Spot is ahaed of where we paused\n")
										break
									} else {
										fmt.Printf("Updating recent clone\n")
										w32.WriteProcessMemory(handle, obj1Loc+uint32(((backDraw-1)%10)*0x1000), replayData[spotToUse].obj1, 0x30)
										w32.WriteProcessMemory(handle, obj2Loc+uint32(((backDraw-1)%10)*0x1000), replayData[spotToUse].obj2, uint(d2RecordSize))
										lastDrawn = lastDrawn + 10
										backDraw--
									}
								} else {
									break
								}
							}
							w32.WriteProcessMemory(handle, d1Loc, replayData[currSpot].obj1, 0x30)
							w32.WriteProcessMemory(handle, d2Loc+uint32(d2RecordOffset), replayData[currSpot].obj2, uint(d2RecordSize))
							fmt.Printf("Spot: %d, Len: %d\n", currSpot, len(replayData))
						} else {
							fmt.Printf("Were at the End of recording\n")
						}

					}
				}
			}
		}
		if button == 3 {
			if pressed == true {
				if menu == 16 {

					fmt.Printf("Old Main: %d\n", oldMainPointer)
					resumeGame()
				}
			}
		}
	} else if saveStates {
		if menu == 16 || menu == 17 {
			if button == 1 {
				if pressed == true {
					currentState++
					if currentState >= 10 {
						currentState = 0
					}
					w32.WriteProcessMemoryAsUint32(handle, 0x0174B050, uint32(currentState))
				}
			}
			if button == 0 {
				if pressed == true {
					currentState--
					if currentState < 0 {
						currentState = 9
					}
					w32.WriteProcessMemoryAsUint32(handle, 0x0174B050, uint32(currentState))
				}
			}
			if button == 6 {
				if pressed == true {
					fmt.Printf("Saving State\n")
					saveState()
				}
			}
			if button == 5 {
				if pressed == true {
					fmt.Printf("Loading State\n")
					loadState()
				}
			}
		}
	}
}

func resumeGame() {
	if !recording {
		distanceFromPause = 0
		lastDrawn = 0
		backDraw = 0
		w32.WriteProcessMemory(handle, 0x0174AFF7, []byte{0}, uint(1))
		w32.WriteProcessMemory(handle, 0x01DE4664, []byte{0}, uint(1))
		emptyData := make([]byte, 0x50-0x8)
		for i := 0; i <= maxClones; i++ {
			w32.WriteProcessMemory(handle, insertLoc+uint32(i*0x1000)+0x8, emptyData, uint(0x50-0x8))
		}
		/*nextAddr, _ := w32.ReadProcessMemoryAsUint32(handle, insertLoc+0x4)
		w32.WriteProcessMemoryAsUint32(handle, loc+0x4, nextAddr)
		w32.WriteProcessMemoryAsUint32(handle, nextAddr, loc)
		*/

		if currSpot < len(replayData) {
			w32.WriteProcessMemory(handle, 0x01DE94A0, replayData[currSpot].grav, uint(12))
			w32.WriteProcessMemory(handle, 0x0174AFDB, replayData[currSpot].time, uint(3))
		}
		replayData = replayData[:currSpot]
		fmt.Printf("Spot: %d, Len: %d\n", currSpot, len(replayData))
		fmt.Printf("Loc: %d, Pointer: %d\n", loc, oldMainPointer)
		w32.WriteProcessMemoryAsUint32(handle, loc+0x10, oldMainPointer)

		recording = true
	}
}
func rewinder() {
	rewinderEnabled = true
	replayData = make([]PlayerData, 0)
	lastMenu := -1
	var oldMili = -1
	currSpot = 0
	jumpDistance = 5
	lastOffsetUsed = 0
	distanceFromPause = 0
	backDraw = 0
	lastDrawn = 0
	recording = true
	go inputHandler(&lastMenu)

	for {
		if rewinderEnabled == false {
			fmt.Printf("Closing Rewinder, recording: %t\n", recording)
			if !recording {
				resumeGame()
			}
			return
		}
		newMenuByte, _ := w32.ReadProcessMemory(handle, 0x01934BE0, uint(1))
		newMenu := int(newMenuByte[0])
		if newMenu != lastMenu {
			fmt.Printf("Meun: %d\n", newMenu)
			lastMenu = newMenu
			if newMenu == 16 {
				fmt.Printf("Setting Loc\n")
				loc, _ = w32.ReadProcessMemoryAsUint32(handle, 0x01DEA6E0)
				d1Loc, _ = w32.ReadProcessMemoryAsUint32(handle, loc+0x34)
				d2Loc, _ = w32.ReadProcessMemoryAsUint32(handle, loc+0x40)
			}
			if newMenu < 4 {
				replayData = make([]PlayerData, 0)
				w32.WriteProcessMemory(handle, 0x0174AFF7, []byte{0}, uint(1))
			}
			if newMenu == 7 {
				//replayData = make([]PlayerData, 0)
				currSpot = 0
				gameTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, uint(3))
				for i := 0; i < len(replayData); i++ {
					if string(replayData[i].time) == string(gameTime) {
						currSpot = i
						replayData = replayData[:currSpot]
						break
					}
				}
				w32.WriteProcessMemory(handle, 0x0174AFF7, []byte{0}, uint(1))
			}

		}
		if lastMenu == 16 {
			currTime, _ := w32.ReadProcessMemory(handle, 0x0174AFDB, 3)
			mili := int(currTime[2])
			if mili != oldMili {
				if recording {

					tempData := new(PlayerData)
					data, _ := w32.ReadProcessMemory(handle, d1Loc, 0x30)
					tempData.obj1 = data
					data, _ = w32.ReadProcessMemory(handle, d2Loc+uint32(d2RecordOffset), uint(d2RecordSize))
					tempData.obj2 = data
					data, _ = w32.ReadProcessMemory(handle, 0x01DE94A0, uint(12))
					tempData.grav = data
					data, _ = w32.ReadProcessMemory(handle, 0x0174AFDB, uint(3))
					tempData.time = data
					replayData = append(replayData, *tempData)
					currSpot++
				} else {

				}
				oldMili = mili
			}

		}
		time.Sleep(1000000)
	}
}
func inputHandler(menu *int) {
	fmt.Printf("Loaded input handler\n")
	oldInputs := make([]uint8, 12*8)
	for {
		newInputs := make([]uint8, 12*8)

		buttons, _ := w32.ReadProcessMemory(handle, 0x01A52C4C, uint(12))

		for i := 0; i < len(newInputs); i++ {
			newInputs[i] = checkByte(buttons[int(i/8)], i%8)
		}

		for i := 0; i < len(newInputs); i++ {
			if newInputs[i] != oldInputs[i] {
				buttonChange(i, !(newInputs[i] == 0), *menu)
			}
		}
		copy(oldInputs, newInputs)
		if rewinderEnabled == false && saveStates == false {
			return
		}
	}
}
