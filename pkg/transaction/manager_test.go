package transaction

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/iceymoss/go-hichat-api/pkg/db"

	"gorm.io/gorm"
)

type Product struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"name"`
	Price float64
	Stock int // 库存量
}

type CartItem struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	ProductID uint
	Quantity  int // 购买数量
}

type Order struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint
	ProductID  uint
	Quantity   int
	TotalPrice float64
}

// CreateOrder 创建订单
func CreateOrder(ctx context.Context, userID uint, productID uint, quantity int) error {
	mysqlConn := GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))

	// 查询商品价格
	var product Product
	if err := mysqlConn.Table("products").Where("id = ?", productID).First(&product).Error; err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// 计算总价
	totalPrice := float64(quantity) * product.Price

	// 创建订单
	order := Order{
		UserID:     userID,
		ProductID:  productID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
	}

	if err := mysqlConn.Table("orders").Create(&order).Error; err != nil {
		return fmt.Errorf("create order failed: %w", err)
	}

	return nil
}

// DeductInventory 扣减库存
func DeductInventory(ctx context.Context, productID uint, quantity int) error {
	mysqlConn := GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))

	// 检查库存
	var product Product
	if err := mysqlConn.Table("products").Where("id = ?", productID).First(&product).Error; err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Stock < quantity {
		return errors.New("insufficient stock")
	}

	// 扣减库存
	if err := mysqlConn.Table("products").
		Where("id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error; err != nil {
		return fmt.Errorf("deduct inventory failed: %w", err)
	}

	return nil
}

// RemoveCartItems 移除购物车商品
func RemoveCartItems(ctx context.Context, userID uint, productID uint) error {
	mysqlConn := GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))

	// 删除购物车项
	if err := mysqlConn.Table("cart_items").
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&CartItem{}).Error; err != nil {
		return fmt.Errorf("remove cart items failed: %w", err)
	}

	return nil
}

func TestTransactionManager(t *testing.T) {
	t.Skip("example transaction flow requires dedicated application fixtures")
	// 初始化事务管理器（实际项目中应该从依赖注入获取）
	txManager := NewManager()

	// 模拟用户ID和商品ID
	userID := uint(1)
	productID := uint(101)
	quantity := 2

	ctx := context.Background()
	err := txManager.Execute(ctx, nil, func(ctx context.Context) error {
		// 1. 扣减库存
		if err := DeductInventory(ctx, productID, quantity); err != nil {
			return fmt.Errorf("deduct inventory error: %w", err)
		}

		// 2. 创建订单
		if err := CreateOrder(ctx, userID, productID, quantity); err != nil {
			return fmt.Errorf("create order error: %w", err)
		}

		// 3. 从购物车移除
		if err := RemoveCartItems(ctx, userID, productID); err != nil {
			return fmt.Errorf("remove cart items error: %w", err)
		}

		return nil
	})
	if err != nil {
		t.Errorf("购买失败 执行事务失败: %s", err.Error())
		return
	}

	t.Log("购买成功!")
}
