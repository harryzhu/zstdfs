package potato

import (
	"encoding/json"
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

func (fo *FileObject) RestatFileObject(stat int) *FileObject {
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

	sb := filobj.RestatFileObject(-1)

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

	sb := filobj.RestatFileObject(0)

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
