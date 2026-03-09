package service

import (
	"errors"

	"itsyourturnring/database"
	"itsyourturnring/model"
)

type AddressService struct{}

func NewAddressService() *AddressService {
	return &AddressService{}
}

// CreateAddress 创建收货地址
func (s *AddressService) CreateAddress(userID int64, address *model.Address) (*model.Address, error) {
	db := database.GetDB()

	// 如果是默认地址，先取消其他默认
	if address.IsDefault {
		_, _ = db.Exec("UPDATE addresses SET is_default = FALSE WHERE user_id = ?", userID)
	}

	result, err := db.Exec(`
		INSERT INTO addresses (user_id, name, phone, province, city, district, detail, is_default)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, address.Name, address.Phone, address.Province, address.City,
		address.District, address.Detail, address.IsDefault)
	if err != nil {
		return nil, err
	}

	addressID, _ := result.LastInsertId()
	return s.GetAddressByID(addressID, userID)
}

// UpdateAddress 更新收货地址
func (s *AddressService) UpdateAddress(addressID int64, userID int64, address *model.Address) (*model.Address, error) {
	db := database.GetDB()

	// 验证地址所属
	var ownerID int64
	err := db.QueryRow("SELECT user_id FROM addresses WHERE id = ?", addressID).Scan(&ownerID)
	if err != nil {
		return nil, errors.New("地址不存在")
	}
	if ownerID != userID {
		return nil, errors.New("无权修改此地址")
	}

	// 如果是默认地址，先取消其他默认
	if address.IsDefault {
		_, _ = db.Exec("UPDATE addresses SET is_default = FALSE WHERE user_id = ?", userID)
	}

	_, err = db.Exec(`
		UPDATE addresses SET name = ?, phone = ?, province = ?, city = ?, district = ?,
			detail = ?, is_default = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		address.Name, address.Phone, address.Province, address.City, address.District,
		address.Detail, address.IsDefault, addressID)
	if err != nil {
		return nil, err
	}

	return s.GetAddressByID(addressID, userID)
}

// DeleteAddress 删除收货地址
func (s *AddressService) DeleteAddress(addressID int64, userID int64) error {
	db := database.GetDB()

	result, err := db.Exec("DELETE FROM addresses WHERE id = ? AND user_id = ?", addressID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("地址不存在或无权删除")
	}

	return nil
}

// GetAddressByID 根据ID获取地址
func (s *AddressService) GetAddressByID(addressID int64, userID int64) (*model.Address, error) {
	db := database.GetDB()

	var address model.Address
	err := db.QueryRow(`
		SELECT id, user_id, name, phone, province, city, district, detail, is_default,
			created_at, updated_at
		FROM addresses WHERE id = ? AND user_id = ?`, addressID, userID).Scan(
		&address.ID, &address.UserID, &address.Name, &address.Phone,
		&address.Province, &address.City, &address.District, &address.Detail,
		&address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

// ListAddresses 地址列表
func (s *AddressService) ListAddresses(userID int64) ([]model.Address, error) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, user_id, name, phone, province, city, district, detail, is_default,
			created_at, updated_at
		FROM addresses WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []model.Address
	for rows.Next() {
		var address model.Address
		err := rows.Scan(
			&address.ID, &address.UserID, &address.Name, &address.Phone,
			&address.Province, &address.City, &address.District, &address.Detail,
			&address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			continue
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetDefaultAddress 获取默认地址
func (s *AddressService) GetDefaultAddress(userID int64) (*model.Address, error) {
	db := database.GetDB()

	var address model.Address
	err := db.QueryRow(`
		SELECT id, user_id, name, phone, province, city, district, detail, is_default,
			created_at, updated_at
		FROM addresses WHERE user_id = ? AND is_default = TRUE`, userID).Scan(
		&address.ID, &address.UserID, &address.Name, &address.Phone,
		&address.Province, &address.City, &address.District, &address.Detail,
		&address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
	if err != nil {
		// 如果没有默认地址，返回第一个
		err = db.QueryRow(`
			SELECT id, user_id, name, phone, province, city, district, detail, is_default,
				created_at, updated_at
			FROM addresses WHERE user_id = ?
			ORDER BY created_at DESC LIMIT 1`, userID).Scan(
			&address.ID, &address.UserID, &address.Name, &address.Phone,
			&address.Province, &address.City, &address.District, &address.Detail,
			&address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &address, nil
}

// SetDefaultAddress 设置默认地址
func (s *AddressService) SetDefaultAddress(addressID int64, userID int64) error {
	db := database.GetDB()

	// 先取消其他默认
	_, _ = db.Exec("UPDATE addresses SET is_default = FALSE WHERE user_id = ?", userID)

	result, err := db.Exec("UPDATE addresses SET is_default = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		addressID, userID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("地址不存在")
	}

	return nil
}
