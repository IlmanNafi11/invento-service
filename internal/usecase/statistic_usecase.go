package usecase

import (
	"context"
	"invento-service/internal/dto"
	"invento-service/internal/rbac"
	"invento-service/internal/usecase/repo"

	"gorm.io/gorm"
)

type StatisticUsecase interface {
	GetStatistics(ctx context.Context, userID, userRole string) (*dto.StatisticData, error)
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

func (su *statisticUsecase) GetStatistics(ctx context.Context, userID, userRole string) (*dto.StatisticData, error) {
	result := &dto.StatisticData{}

	hasProjectRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Project", "read")
	hasModulRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Modul", "read")
	hasUserRead, _ := su.casbinEnforcer.CheckPermission(userRole, "User", "read")
	hasRoleRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Role", "read")

	if !hasProjectRead && !hasModulRead && !hasUserRead && !hasRoleRead {
		return result, nil
	}

	type statisticCounts struct {
		TotalProject int64 `gorm:"column:total_project"`
		TotalModul   int64 `gorm:"column:total_modul"`
		TotalUser    int64 `gorm:"column:total_user"`
		TotalRole    int64 `gorm:"column:total_role"`
	}

	var counts statisticCounts
	su.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM projects WHERE user_id = ?) AS total_project,
			(SELECT COUNT(*) FROM moduls WHERE user_id = ?) AS total_modul,
			(SELECT COUNT(*) FROM user_profiles) AS total_user,
			(SELECT COUNT(*) FROM roles) AS total_role
	`, userID, userID).Scan(&counts)

	if hasProjectRead {
		count := int(counts.TotalProject)
		result.TotalProject = &count
	}
	if hasModulRead {
		count := int(counts.TotalModul)
		result.TotalModul = &count
	}
	if hasUserRead {
		count := int(counts.TotalUser)
		result.TotalUser = &count
	}
	if hasRoleRead {
		count := int(counts.TotalRole)
		result.TotalRole = &count
	}

	return result, nil
}
