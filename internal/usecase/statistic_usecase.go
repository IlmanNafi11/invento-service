package usecase

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"gorm.io/gorm"
)

type StatisticUsecase interface {
	GetStatistics(userID uint, userRole string) (*domain.StatisticData, error)
}

type statisticUsecase struct {
	userRepo       repo.UserRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	roleRepo       repo.RoleRepository
	casbinEnforcer *helper.CasbinEnforcer
	db             *gorm.DB
}

func NewStatisticUsecase(
	userRepo repo.UserRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	roleRepo repo.RoleRepository,
	casbinEnforcer *helper.CasbinEnforcer,
	db *gorm.DB,
) StatisticUsecase {
	return &statisticUsecase{
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		modulRepo:      modulRepo,
		roleRepo:       roleRepo,
		casbinEnforcer: casbinEnforcer,
		db:             db,
	}
}

func (su *statisticUsecase) GetStatistics(userID uint, userRole string) (*domain.StatisticData, error) {
	result := &domain.StatisticData{}

	hasProjectRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Project", "read")
	hasModulRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Modul", "read")
	hasUserRead, _ := su.casbinEnforcer.CheckPermission(userRole, "User", "read")
	hasRoleRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Role", "read")

	if hasProjectRead {
		projectCount, _ := su.projectRepo.CountByUserID(userID)
		result.TotalProject = &projectCount
	}

	if hasModulRead {
		modulCount, _ := su.modulRepo.CountByUserID(userID)
		result.TotalModul = &modulCount
	}

	if hasUserRead {
		var totalUser int64
		su.db.Model(&domain.User{}).Count(&totalUser)
		count := int(totalUser)
		result.TotalUser = &count
	}

	if hasRoleRead {
		var totalRole int64
		su.db.Model(&domain.Role{}).Count(&totalRole)
		count := int(totalRole)
		result.TotalRole = &count
	}

	return result, nil
}
