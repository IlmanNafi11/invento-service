package domain

import (
	"invento-service/internal/dto"
	"testing"
)

func TestStatisticData(t *testing.T) {
	t.Parallel()
	t.Run("dto.StatisticData with all fields", func(t *testing.T) {
		t.Parallel()
		totalProject := 100
		totalModul := 500
		totalUser := 25
		totalRole := 5

		data := dto.StatisticData{
			TotalProject: &totalProject,
			TotalModul:   &totalModul,
			TotalUser:    &totalUser,
			TotalRole:    &totalRole,
		}

		if data.TotalProject == nil || *data.TotalProject != 100 {
			t.Error("Expected TotalProject to be 100")
		}
		if data.TotalModul == nil || *data.TotalModul != 500 {
			t.Error("Expected TotalModul to be 500")
		}
		if data.TotalUser == nil || *data.TotalUser != 25 {
			t.Error("Expected TotalUser to be 25")
		}
		if data.TotalRole == nil || *data.TotalRole != 5 {
			t.Error("Expected TotalRole to be 5")
		}
	})

	t.Run("dto.StatisticData with partial fields", func(t *testing.T) {
		t.Parallel()
		totalProject := 50

		data := dto.StatisticData{
			TotalProject: &totalProject,
		}

		if data.TotalProject == nil || *data.TotalProject != 50 {
			t.Error("Expected TotalProject to be 50")
		}
		if data.TotalModul != nil {
			t.Error("Expected TotalModul to be nil")
		}
		if data.TotalUser != nil {
			t.Error("Expected TotalUser to be nil")
		}
		if data.TotalRole != nil {
			t.Error("Expected TotalRole to be nil")
		}
	})

	t.Run("dto.StatisticData with zero values", func(t *testing.T) {
		t.Parallel()
		totalProject := 0
		totalModul := 0
		totalUser := 0
		totalRole := 0

		data := dto.StatisticData{
			TotalProject: &totalProject,
			TotalModul:   &totalModul,
			TotalUser:    &totalUser,
			TotalRole:    &totalRole,
		}

		if data.TotalProject == nil || *data.TotalProject != 0 {
			t.Error("Expected TotalProject to be 0")
		}
		if data.TotalModul == nil || *data.TotalModul != 0 {
			t.Error("Expected TotalModul to be 0")
		}
		if data.TotalUser == nil || *data.TotalUser != 0 {
			t.Error("Expected TotalUser to be 0")
		}
		if data.TotalRole == nil || *data.TotalRole != 0 {
			t.Error("Expected TotalRole to be 0")
		}
	})

	t.Run("dto.StatisticData all nil", func(t *testing.T) {
		t.Parallel()
		data := dto.StatisticData{}

		if data.TotalProject != nil {
			t.Error("Expected TotalProject to be nil")
		}
		if data.TotalModul != nil {
			t.Error("Expected TotalModul to be nil")
		}
		if data.TotalUser != nil {
			t.Error("Expected TotalUser to be nil")
		}
		if data.TotalRole != nil {
			t.Error("Expected TotalRole to be nil")
		}
	})
}

func TestStatisticResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.StatisticResponse with complete data", func(t *testing.T) {
		t.Parallel()
		totalProject := 150
		totalModul := 750
		totalUser := 30
		totalRole := 6

		resp := dto.StatisticResponse{
			Data: dto.StatisticData{
				TotalProject: &totalProject,
				TotalModul:   &totalModul,
				TotalUser:    &totalUser,
				TotalRole:    &totalRole,
			},
		}

		if resp.Data.TotalProject == nil || *resp.Data.TotalProject != 150 {
			t.Error("Expected TotalProject to be 150")
		}
		if resp.Data.TotalModul == nil || *resp.Data.TotalModul != 750 {
			t.Error("Expected TotalModul to be 750")
		}
		if resp.Data.TotalUser == nil || *resp.Data.TotalUser != 30 {
			t.Error("Expected TotalUser to be 30")
		}
		if resp.Data.TotalRole == nil || *resp.Data.TotalRole != 6 {
			t.Error("Expected TotalRole to be 6")
		}
	})

	t.Run("dto.StatisticResponse with partial data", func(t *testing.T) {
		t.Parallel()
		totalProject := 200

		resp := dto.StatisticResponse{
			Data: dto.StatisticData{
				TotalProject: &totalProject,
			},
		}

		if resp.Data.TotalProject == nil || *resp.Data.TotalProject != 200 {
			t.Error("Expected TotalProject to be 200")
		}
		if resp.Data.TotalModul != nil {
			t.Error("Expected TotalModul to be nil")
		}
	})

	t.Run("dto.StatisticResponse with empty data", func(t *testing.T) {
		t.Parallel()
		resp := dto.StatisticResponse{
			Data: dto.StatisticData{},
		}

		if resp.Data.TotalProject != nil {
			t.Error("Expected TotalProject to be nil")
		}
		if resp.Data.TotalModul != nil {
			t.Error("Expected TotalModul to be nil")
		}
		if resp.Data.TotalUser != nil {
			t.Error("Expected TotalUser to be nil")
		}
		if resp.Data.TotalRole != nil {
			t.Error("Expected TotalRole to be nil")
		}
	})
}

func TestStatisticDataMutability(t *testing.T) {
	t.Parallel()
	t.Run("Modify dto.StatisticData pointers", func(t *testing.T) {
		t.Parallel()
		totalProject := 100
		totalModul := 200

		data := dto.StatisticData{
			TotalProject: &totalProject,
			TotalModul:   &totalModul,
		}

		if *data.TotalProject != 100 {
			t.Error("Expected TotalProject to be 100")
		}

		newTotal := 150
		data.TotalProject = &newTotal

		if *data.TotalProject != 150 {
			t.Error("Expected TotalProject to be updated to 150")
		}
		if *data.TotalModul != 200 {
			t.Error("Expected TotalModul to remain 200")
		}
	})
}

func TestStatisticDataEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Large numbers", func(t *testing.T) {
		t.Parallel()
		totalProject := 999999
		totalModul := 1000000
		totalUser := 50000
		totalRole := 100

		data := dto.StatisticData{
			TotalProject: &totalProject,
			TotalModul:   &totalModul,
			TotalUser:    &totalUser,
			TotalRole:    &totalRole,
		}

		if *data.TotalProject != 999999 {
			t.Error("Expected TotalProject to be 999999")
		}
		if *data.TotalModul != 1000000 {
			t.Error("Expected TotalModul to be 1000000")
		}
		if *data.TotalUser != 50000 {
			t.Error("Expected TotalUser to be 50000")
		}
		if *data.TotalRole != 100 {
			t.Error("Expected TotalRole to be 100")
		}
	})

	t.Run("Single field set", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name   string
			field  string
			value  int
			setter func(*dto.StatisticData, *int)
		}{
			{
				name:   "TotalProject",
				value:  10,
				setter: func(d *dto.StatisticData, v *int) { d.TotalProject = v },
			},
			{
				name:   "TotalModul",
				value:  20,
				setter: func(d *dto.StatisticData, v *int) { d.TotalModul = v },
			},
			{
				name:   "TotalUser",
				value:  5,
				setter: func(d *dto.StatisticData, v *int) { d.TotalUser = v },
			},
			{
				name:   "TotalRole",
				value:  3,
				setter: func(d *dto.StatisticData, v *int) { d.TotalRole = v },
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				value := tc.value
				data := &dto.StatisticData{}
				tc.setter(data, &value)

				// Verify the field is set
				switch tc.name {
				case "TotalProject":
					if data.TotalProject == nil || *data.TotalProject != tc.value {
						t.Errorf("Expected %s to be %d", tc.name, tc.value)
					}
				case "TotalModul":
					if data.TotalModul == nil || *data.TotalModul != tc.value {
						t.Errorf("Expected %s to be %d", tc.name, tc.value)
					}
				case "TotalUser":
					if data.TotalUser == nil || *data.TotalUser != tc.value {
						t.Errorf("Expected %s to be %d", tc.name, tc.value)
					}
				case "TotalRole":
					if data.TotalRole == nil || *data.TotalRole != tc.value {
						t.Errorf("Expected %s to be %d", tc.name, tc.value)
					}
				}
			})
		}
	})
}
