package usecase

import (
	"context"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/rbac"
	"invento-service/internal/usecase/repo"

	"gorm.io/gorm"
)

type StatisticUsecase interface {
	GetStatistics(ctx context.Context, userID string, userRole string) (*dto.StatisticData, error)
}

type statisticUsecase struct {
	userRepo       repo.UserRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	roleRepo       repo.RoleRepository
	casbinEnforcer *rbac.CasbinEnforcer
	db             *gorm.DB
}

func NewStatisticUsecase(
	userRepo repo.UserRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	roleRepo repo.RoleRepository,
	casbinEnforcer *rbac.CasbinEnforcer,
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

func (su *statisticUsecase) GetStatistics(ctx context.Context, userID string, userRole string) (*dto.StatisticData, error) {
	result := &dto.StatisticData{}

	hasProjectRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Project", "read")
	hasModulRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Modul", "read")
	hasUserRead, _ := su.casbinEnforcer.CheckPermission(userRole, "User", "read")
	hasRoleRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Role", "read")

	if hasProjectRead {
		projectCount, _ := su.projectRepo.CountByUserID(ctx, userID)
		result.TotalProject = &projectCount
	}

	if hasModulRead {
		modulCount, _ := su.modulRepo.CountByUserID(ctx, userID)
		result.TotalModul = &modulCount
	}

	if hasUserRead {
		var totalUser int64
		su.db.WithContext(ctx).Model(&domain.User{}).Count(&totalUser)
		count := int(totalUser)
		result.TotalUser = &count
	}

	if hasRoleRead {
		var totalRole int64
		su.db.WithContext(ctx).Model(&domain.Role{}).Count(&totalRole)
		count := int(totalRole)
		result.TotalRole = &count
	}

	return result, nil
}
