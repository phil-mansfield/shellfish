package io

import (
	"io"
	"encoding/binary"
	"fmt"

	"unsafe"
)

type FilterData struct {
	Snaps []int64
	IDs []int64
	Particles [][]int64
}

type haloInfo struct {
	id, snap, startByte, len int64
}

func flagToOrder(orderFlag string) binary.ByteOrder {
	switch orderFlag {
	case "LittleEndian":
		return binary.LittleEndian
	case "BigEndian":
		return binary.BigEndian
	case "SystemOrder":
		if IsSysOrder(binary.BigEndian) {
			return binary.BigEndian
		}
		return binary.LittleEndian
	}
	panic(fmt.Sprintf("Unknown orderFlag, '%s'", orderFlag))
}

func WriteFilter(wr io.Writer, orderFlag string, data FilterData) error {
	order := flagToOrder(orderFlag)

	if order == binary.BigEndian {
		binary.Write(wr, order, int32(-1))
	} else {
		binary.Write(wr, order, int32(0))
	}
	binary.Write(wr, order, int32(len(data.Snaps)))

	info := make([]haloInfo, len(data.Snaps))
	baseOffset := int64(4 + 4 + len(info) * int(unsafe.Sizeof(haloInfo{})))

	totLen := 0
	for i := range info {
		info[i].snap = data.Snaps[i]
		info[i].id = data.IDs[i]
		info[i].startByte = int64(totLen * 8) + baseOffset
		info[i].len = int64(len(data.Particles[i]))
	}
	binary.Write(wr, order, info)

	for i := range info {
		binary.Write(wr, order, data.Particles[i])
	}

	return nil
}

func ReadFilter(rd io.Reader) (FilterData, error) {
	var orderFlag int32
	binary.Read(rd, binary.LittleEndian, &orderFlag)

	var order binary.ByteOrder
	switch orderFlag {
	case 0:
		order = binary.LittleEndian
	case -1:
		order = binary.BigEndian
	default:
		return FilterData{}, fmt.Errorf(
			"Unknown endianness flag at start of file.")
	}

	var haloNum int32
	binary.Read(rd, order, &haloNum)

	info := make([]haloInfo, haloNum)
	binary.Read(rd, order, info)

	data := FilterData{}
	data.Snaps = make([]int64, haloNum)
	data.IDs = make([]int64, haloNum)
	data.Particles = make([][]int64, haloNum)

	for i := range info {
		data.Snaps[i] = info[i].snap
		data.IDs[i] = info[i].id
		data.Particles[i] = make([]int64, info[i].len)
		binary.Read(rd, order, data.Particles[i])
	}

	return data, nil
}