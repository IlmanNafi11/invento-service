package helper

import (
	"fiber-boiler-plate/internal/usecase/repo"
	"log"
	"time"
)

type RefreshTokenCleanup struct {
	refreshTokenRepo repo.RefreshTokenRepository
	cleanupInterval  time.Duration
	stopChan         chan bool
}

func NewRefreshTokenCleanup(refreshTokenRepo repo.RefreshTokenRepository, intervalHours int) *RefreshTokenCleanup {
	return &RefreshTokenCleanup{
		refreshTokenRepo: refreshTokenRepo,
		cleanupInterval:  time.Duration(intervalHours) * time.Hour,
		stopChan:         make(chan bool),
	}
}

func (rtc *RefreshTokenCleanup) Start() {
	ticker := time.NewTicker(rtc.cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				rtc.cleanup()
			case <-rtc.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	log.Println("RefreshTokenCleanup: background cleanup dimulai dengan interval", rtc.cleanupInterval)
}

func (rtc *RefreshTokenCleanup) Stop() {
	rtc.stopChan <- true
	log.Println("RefreshTokenCleanup: background cleanup dihentikan")
}

func (rtc *RefreshTokenCleanup) cleanup() {
	if err := rtc.refreshTokenRepo.CleanupExpired(); err != nil {
		log.Printf("RefreshTokenCleanup: error saat cleanup - %v", err)
		return
	}
	log.Println("RefreshTokenCleanup: berhasil membersihkan refresh token yang expired")
}
