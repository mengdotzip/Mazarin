package database

import (
	"log"
	"net"
	"sync"
)

//have hashing here in the future

type UserInfo struct {
	Name     string
	Ip       string
	Sessions []net.Conn
}

var dbMutex = sync.RWMutex{}
var usersMapName = make(map[string]*UserInfo)
var usersMapIp = make(map[string][]*UserInfo)

func GetUserByName(name string) *UserInfo {
	dbMutex.RLock()
	info := usersMapName[name]
	dbMutex.RUnlock()
	return info
}

func GetUserByIp(ip string) []*UserInfo {
	dbMutex.RLock()
	info := usersMapIp[ip]
	dbMutex.RUnlock()
	return info

}

func InsertUser(info UserInfo) bool {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	_, ok := usersMapIp[info.Ip]
	if ok {
		log.Println("ERROR db: user already exists")
		return false
	}

	_, okN := usersMapName[info.Name]
	if okN {
		log.Println("ERROR db: user already exists")
		return false
	}

	usersMapIp[info.Ip] = append(usersMapIp[info.Ip], &info)
	usersMapName[info.Name] = &info

	return true
}

func DeleteUser(name string) bool {

	return true
}
