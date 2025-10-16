package helper

import (
	"log"
	"time"
)

type KeyRotationScheduler struct {
	jwtManager       *JWTManager
	rotationInterval time.Duration
	stopChan         chan bool
}

func NewKeyRotationScheduler(jwtManager *JWTManager, intervalHours int) *KeyRotationScheduler {
	return &KeyRotationScheduler{
		jwtManager:       jwtManager,
		rotationInterval: time.Duration(intervalHours) * time.Hour,
		stopChan:         make(chan bool),
	}
}

func (krs *KeyRotationScheduler) Start() {
	ticker := time.NewTicker(krs.rotationInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				krs.rotateKeys()
			case <-krs.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	log.Println("KeyRotation: scheduler dimulai dengan interval", krs.rotationInterval)
}

func (krs *KeyRotationScheduler) Stop() {
	krs.stopChan <- true
	log.Println("KeyRotation: scheduler dihentikan")
}

func (krs *KeyRotationScheduler) rotateKeys() {
	krs.jwtManager.RotateKeys()
	log.Printf("KeyRotation: kunci berhasil dirotasi ke %s", krs.jwtManager.GetKeyID())
}
