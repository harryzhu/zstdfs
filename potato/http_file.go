package potato

import (
	"encoding/json"
	//"errors"
)

type FileObject struct {
	Ver  int
	Stat int
	Csec string
	Msec string
	Name string
	Mime string
	Size string
	Data []byte
}

type FileResponse struct {
	URL  string
	Name string
	Mime string
	Size string
}

func (fo *FileObject) reStatFileObject(stat int) *FileObject {
	if fo.Ver >= 0 {
		fo.Ver += 1
	} else {
		fo.Ver = 0
	}
	fo.Msec = TimeNowUnixString()
	fo.Stat = stat
	return fo
}

func FileBan(key []byte) error {
	if EntityExists(key) == false {
		return nil
	}

	data, err := EntityGet(key)

	if err != nil {
		return err
	}

	var filobj FileObject
	err = json.Unmarshal(data, &filobj)

	if err != nil {
		return err
	}

	sb := filobj.reStatFileObject(-1)

	byteFileObject, err := json.Marshal(sb)

	if err != nil {
		return err
	}

	err = EntitySet(key, byteFileObject)

	if err != nil {
		return err
	}

	return nil
}

func FilePub(key []byte) error {
	if EntityExists(key) == false {
		return nil
	}

	data, err := EntityGet(key)

	if err != nil {
		return err
	}

	var filobj FileObject
	err = json.Unmarshal(data, &filobj)

	if err != nil {
		return err
	}

	sb := filobj.reStatFileObject(0)

	byteFileObject, err := json.Marshal(sb)

	if err != nil {
		return err
	}

	err = EntitySet(key, byteFileObject)

	if err != nil {
		return err
	}

	return nil
}

func FileHead(key []byte) ([]byte, error) {
	if EntityExists(key) == false {
		return nil, ErrKeyNotFound
	}

	data, err := EntityGet(key)

	if err != nil {
		return nil, err
	}

	var filobj FileObject
	err = json.Unmarshal(data, &filobj)

	if err != nil {
		return nil, err
	}

	filobj.Data = nil

	byteFileObject, err := json.Marshal(filobj)

	if err != nil {
		return nil, err
	}

	return byteFileObject, nil
}
