package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationrequest "foundation/foundation-request"
	foundationuser "foundation/foundation-user"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	metastring "meta/meta-string"
	"meta/singleton"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

var singletonUserDao = singleton.Singleton[UserDao]{}

// GetDB 获取数据库连接
func (dao *UserDao) GetDB(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}

// WithTransaction 在事务中执行操作
func (dao *UserDao) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := dao.GetDB(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetUserDao() *UserDao {
	return singletonUserDao.GetInstance(
		func() *UserDao {
			dao := &UserDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *UserDao) GetUserLogin(ctx context.Context, id int) (*foundationview.UserLogin, error) {
	var user foundationview.UserLogin
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, password").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user login info")
	}
	return &user, nil
}

func (d *UserDao) GetUserLoginByUsername(ctx context.Context, username string) (*foundationview.UserLogin, error) {
	var user foundationview.UserLogin
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, password").
		Where("LOWER(username) = LOWER(?)", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user login info by username")
	}
	return &user, nil
}

func (d *UserDao) GetUserPassword(ctx context.Context, userId int) (string, error) {
	var password string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("password").
		Where("id = ?", userId).
		Pluck("password", &password).Error
	if err != nil {
		return "", metaerror.Wrap(err, "get user password")
	}
	return password, nil
}

func (d *UserDao) GetModifyInfo(ctx context.Context, userId int) (*foundationview.UserModifyInfo, error) {
	var userInfo foundationview.UserModifyInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select(
			`id, username, nickname, real_name, 
email, gender, number, slogan, organization, qq, blog,
vjudge_id, github, codeforces`,
		).
		Where("id = ?", userId).
		First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user modify info")
	}
	return &userInfo, nil
}

func (d *UserDao) GetInfoByUsername(ctx context.Context, username string) (*foundationview.UserInfo, error) {
	var userInfo foundationview.UserInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select(
			`id, username, nickname, real_name, 
email, gender, number, slogan, organization, qq, blog,
			vjudge_id, github, codeforces, 
check_in_count, insert_time, modify_time, accept, attempt,
			level, experience, coin`,
		).
		Where("LOWER(username) = LOWER(?)", username).
		First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user info by username")
	}
	return &userInfo, nil
}

func (d *UserDao) GetUserAccountInfo(ctx context.Context, id int) (*foundationview.UserAccountInfo, error) {
	var userAccountInfo foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("id = ?", id).
		Take(&userAccountInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return &userAccountInfo, nil
}

func (d *UserDao) GetUserAccountInfos(ctx context.Context, ids []int) ([]*foundationview.UserAccountInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var userAccountInfos []*foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("id IN ?", ids).
		Find(&userAccountInfos).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return userAccountInfos, nil
}

func (d *UserDao) GetUserAccountInfosByUsername(
	ctx context.Context,
	usernames []string,
) ([]*foundationview.UserAccountInfo, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var userAccountInfos []*foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("LOWER(username) IN ?", metastring.LowerSlice(usernames)).
		Find(&userAccountInfos).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return userAccountInfos, nil
}

func (d *UserDao) GetUserIdByUsername(ctx context.Context, username string) (int, error) {
	var userId int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("LOWER(username) = LOWER(?)", username).
		Pluck("id", &userId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // User not found
		}
		return 0, metaerror.Wrap(err, "get user id by username")
	}
	return userId, nil
}

func (d *UserDao) GetUserIdsByUsername(ctx context.Context, usernames []string) ([]int, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var userIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("LOWER(username) IN ?", metastring.LowerSlice(usernames)).
		Pluck("id", &userIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user ids by username")
	}
	return userIds, nil
}

func (d *UserDao) GetEmail(ctx context.Context, id int) (*string, error) {
	var email string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("id = ?", id).
		Pluck("email", &email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get email by user id")
	}
	return &email, nil
}

func (d *UserDao) GetEmailByUsername(ctx context.Context, username string) (*string, error) {
	var email string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("LOWER(username) = LOWER(?)", username).
		Pluck("email", &email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get email by username")
	}
	return &email, nil
}

func (d *UserDao) GetRankAcAll(ctx context.Context, page int, size int) ([]*foundationview.UserRank, int, error) {
	var userRanks []*foundationview.UserRank
	var total int64
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("accept > 0").
		Count(&total).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "count total users")
	}
	err = d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, email, slogan, accept, attempt").
		Where("accept > 0").
		Order("accept DESC, attempt ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&userRanks).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "get user ranks")
	}
	return userRanks, int(total), nil
}

func (d *UserDao) FilterValidUserIds(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var validIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("id IN ?", ids).
		Pluck("id", &validIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "filter valid user ids")
	}
	return validIds, nil
}

func (d *UserDao) UpdatePassword(ctx context.Context, username string, encodePassword string, nowTime time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("LOWER(username) = LOWER(?)", username).
		Update("password", encodePassword).
		Update("modify_time", nowTime)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user password")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserInfo(
	ctx context.Context,
	userId int,
	request *foundationrequest.UserModifyInfo,
	modifyTime time.Time,
) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})

	gender := foundationenum.GetUserGender(request.Gender)

	res := db.Where("id = ?", userId).
		Updates(
			map[string]interface{}{
				"nickname":     request.Nickname,
				"slogan":       request.Slogan,
				"real_name":    request.RealName,
				"gender":       gender,
				"organization": request.Organization,
				"blog":         request.Blog,
				"modify_time":  modifyTime,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user info")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) InsertUser(ctx context.Context, user *foundationmodel.User) error {
	if user == nil {
		return metaerror.New("user is nil")
	}
	db := d.db.WithContext(ctx).Model(user)
	if err := db.Create(user).Error; err != nil {
		return metaerror.Wrap(err, "insert user")
	}
	return nil
}

func (d *UserDao) UpdatePasswordByUserId(ctx *gin.Context, id int, encodePassword string, nowTime time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Update("password", encodePassword).
		Update("modify_time", nowTime)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user password")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserVjudgeUsername(ctx *gin.Context, id int, vjudgeId string, now time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Updates(
			map[string]interface{}{
				"vjudge_id":   vjudgeId,
				"modify_time": now,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user vjudge username")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserEmail(ctx context.Context, id int, email string, now time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Updates(
			map[string]interface{}{
				"email":       email,
				"modify_time": now,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user email")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) PostLoginLog(ctx context.Context, userId int, nowTime time.Time, ip string, agent string) error {
	loginLog := foundationmodel.NewUserLoginBuilder().
		UserId(userId).InsertTime(nowTime).IP(ip).UserAgent(agent).Build()
	db := d.db.WithContext(ctx).Model(loginLog)
	if err := db.Create(loginLog).Error; err != nil {
		return metaerror.Wrap(err, "insert user login log")
	}
	return nil
}

// AddUserExperienceWithTx 在事务中更新用户经验值并自动更新等级
// 使用数据库行级锁确保在高并发环境下的数据一致性
func (d *UserDao) AddUserExperienceWithTx(tx *gorm.DB, userId int, expGain int, nowTime time.Time) (int, int, error) {
	// 使用FOR UPDATE获取行级锁，确保其他事务无法同时修改该记录
	var user foundationmodel.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Model(&foundationmodel.User{}).
		Where("id = ?", userId).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, metaerror.New("用户不存在")
		}
		return 0, 0, metaerror.Wrap(err, "获取用户信息失败")
	}

	// 计算更新后的经验值
	updatedExp := user.Experience + expGain

	// 计算新等级（使用简化公式：等级 = 经验值 / 100）
	newLevel := foundationuser.GetLevelByExperience(updatedExp)

	// 更新用户经验值和等级
	res := tx.Model(&foundationmodel.User{}).
		Where("id = ?", userId). // 由于已获取行级锁，不需要额外的乐观锁条件
		Updates(map[string]interface{}{
			"experience":  updatedExp,
			"level":       newLevel,
			"modify_time": nowTime,
		})

	if res.Error != nil {
		return 0, 0, metaerror.Wrap(res.Error, "更新用户经验值和等级失败")
	}

	return newLevel, updatedExp, nil
}

// GetUserExperience 获取用户经验值
func (d *UserDao) GetUserExperience(ctx context.Context, userId int) (int, error) {
	var exp int
	err := d.db.WithContext(ctx).Model(&foundationmodel.User{}).Where("id = ?", userId).Pluck("experience", &exp).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "获取用户经验值失败")
	}
	return exp, nil
}

// GetUserExperienceWithTx 在事务中获取用户经验值
func (d *UserDao) GetUserExperienceWithTx(tx *gorm.DB, userId int) (int, error) {
	var exp int
	err := tx.Model(&foundationmodel.User{}).Where("id = ?", userId).Pluck("experience", &exp).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "在事务中获取用户经验值失败")
	}
	return exp, nil
}

// UpdateUserLevel 更新用户等级
func (d *UserDao) UpdateUserLevel(ctx context.Context, userId int, level int) error {
	res := d.db.WithContext(ctx).Model(&foundationmodel.User{}).Where("id = ?", userId).Update("level", level)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "更新用户等级失败")
	}
	return nil
}

// UpdateUserLevelWithTx 在事务中更新用户等级
func (d *UserDao) UpdateUserLevelWithTx(tx *gorm.DB, userId int, level int) error {
	res := tx.Model(&foundationmodel.User{}).Where("id = ?", userId).Update("level", level)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "在事务中更新用户等级失败")
	}
	return nil
}

// InsertUserExperience 插入用户经验记录
func (d *UserDao) InsertUserExperience(ctx context.Context, expRecord *foundationmodel.UserExperience) error {
	if expRecord == nil {
		return metaerror.New("experience record is nil")
	}
	if err := d.db.WithContext(ctx).Create(expRecord).Error; err != nil {
		return metaerror.Wrap(err, "insert user experience record")
	}
	return nil
}

// InsertUserExperienceWithTx 在事务中插入用户经验记录
func (d *UserDao) InsertUserExperienceWithTx(tx *gorm.DB, expRecord *foundationmodel.UserExperience) error {
	if expRecord == nil {
		return metaerror.New("experience record is nil")
	}
	if err := tx.Create(expRecord).Error; err != nil {
		return metaerror.Wrap(err, "insert user experience record in transaction")
	}
	return nil
}

// AddUserCoinWithTx 在事务中更新用户金币
// 使用数据库行级锁确保在高并发环境下的数据一致性
func (d *UserDao) AddUserCoinWithTx(tx *gorm.DB, userId int, coinGain int, nowTime time.Time) (int, error) {
	// 检查用户是否存在
	var existingCoin int
	if err := tx.Model(&foundationmodel.User{}).Where("id = ?", userId).Pluck("coin", &existingCoin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, metaerror.New("用户不存在")
		}
		return 0, metaerror.Wrap(err, "获取用户金币失败")
	}

	// 检查金币是否会变为负数
	if existingCoin+coinGain < 0 {
		return 0, metaerror.New("金币不足")
	}

	// 直接在SQL中使用+操作符更新金币值
	res := tx.Model(&foundationmodel.User{}).
		Where("id = ?", userId).
		Update("coin", gorm.Expr("coin + ?", coinGain)).
		Update("modify_time", nowTime)

	if res.Error != nil {
		return 0, metaerror.Wrap(res.Error, "更新用户金币失败")
	}

	// 返回更新后的金币值
	return existingCoin + coinGain, nil
}

// GetUserCoin 获取用户金币
func (d *UserDao) GetUserCoin(ctx context.Context, userId int) (int, error) {
	var coin int
	err := d.db.WithContext(ctx).Model(&foundationmodel.User{}).Where("id = ?", userId).Pluck("coin", &coin).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "获取用户金币失败")
	}
	return coin, nil
}

// GetUserCoinWithTx 在事务中获取用户金币
func (d *UserDao) GetUserCoinWithTx(tx *gorm.DB, userId int) (int, error) {
	var coin int
	err := tx.Model(&foundationmodel.User{}).Where("id = ?", userId).Pluck("coin", &coin).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "在事务中获取用户金币失败")
	}
	return coin, nil
}

// InsertUserCoin 插入用户金币记录
func (d *UserDao) InsertUserCoin(ctx context.Context, coinRecord *foundationmodel.UserCoin) error {
	if coinRecord == nil {
		return metaerror.New("coin record is nil")
	}
	if err := d.db.WithContext(ctx).Create(coinRecord).Error; err != nil {
		return metaerror.Wrap(err, "insert user coin record")
	}
	return nil
}

// InsertUserCoinWithTx 在事务中插入用户金币记录
func (d *UserDao) InsertUserCoinWithTx(tx *gorm.DB, coinRecord *foundationmodel.UserCoin) error {
	if coinRecord == nil {
		return metaerror.New("coin record is nil")
	}
	if err := tx.Create(coinRecord).Error; err != nil {
		return metaerror.Wrap(err, "insert user coin record in transaction")
	}
	return nil
}

// CheckUserExperienceExists 检查用户经验记录是否存在（基于user_id, type, param的唯一约束）
func (d *UserDao) CheckUserExperienceExists(ctx context.Context, userId int, expType foundationuser.ExperienceType, param string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&foundationmodel.UserExperience{}).
		Where("user_id = ? AND type = ? AND param = ?", userId, expType, param).
		Count(&count).Error
	if err != nil {
		return false, metaerror.Wrap(err, "check user experience exists")
	}
	return count > 0, nil
}

// GetUserUnrewardedACProblems 获取用户尚未获得经验的AC题目
func (d *UserDao) GetUserUnrewardedACProblems(ctx context.Context, userId int) ([]*foundationview.ProblemViewKey, error) {
	var results []struct {
		ProblemId  int    `gorm:"column:problem_id"`
		ProblemKey string `gorm:"column:key"`
	}

	// 使用子查询获取用户已AC的题目ID
	acProblemIds := d.db.Table("judge_job").
		Select("DISTINCT problem_id").
		Where("inserter = ? AND status = ?", userId, foundationjudge.JudgeStatusAC)

	// 使用LEFT JOIN查询用户已AC但未获得经验的题目
	err := d.db.WithContext(ctx).
		Table("problem").
		Select("id AS problem_id, key").
		Where("id IN (?)", acProblemIds).
		Joins("LEFT JOIN user_experience ON user_experience.user_id = ? AND user_experience.type = ? AND user_experience.param = CAST(problem.id AS VARCHAR)", userId, foundationuser.ExperienceTypeAccepted).
		Where("user_experience.user_id IS NULL").
		Scan(&results).Error

	if err != nil {
		return nil, metaerror.Wrap(err, "get user unrewarded ac problems")
	}

	// 转换为ProblemViewKey结构
	var problems []*foundationview.ProblemViewKey
	for _, r := range results {
		problems = append(problems, &foundationview.ProblemViewKey{
			Id:  r.ProblemId,
			Key: r.ProblemKey,
		})
	}

	return problems, nil
}

// AddUserAcceptedExperience 为用户添加奖励经验值（按问题，带防重复检查）
func (d *UserDao) AddUserAcceptedExperience(ctx context.Context, userId int, problemId int, expGain int, nowTime time.Time) (bool, int, int, error) {
	// 检查用户是否已经领取过该问题的奖励（基于唯一约束）
	param := fmt.Sprintf("%d", problemId)
	exists, err := d.CheckUserExperienceExists(ctx, userId, foundationuser.ExperienceTypeAccepted, param)
	if err != nil {
		return false, 0, 0, metaerror.Wrap(err, "检查用户奖励记录失败")
	}
	if exists {
		return true, 0, 0, nil
	}

	var newLevel, newExp int
	// 使用事务确保数据一致性
	err = d.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 插入经验记录
		expRecord := foundationmodel.NewUserExperienceBuilder().
			UserId(userId).
			Value(expGain).
			Type(foundationuser.ExperienceTypeAccepted).
			Param(param).
			InserterTime(nowTime).
			Build()
		insertErr := d.InsertUserExperienceWithTx(tx, expRecord)
		if insertErr != nil {
			// 判断是否是重复插入错误
			if errors.Is(insertErr, gorm.ErrDuplicatedKey) {
				return nil
			}
			return insertErr
		}

		// 更新用户经验值和等级
		var updateErr error
		newLevel, newExp, updateErr = d.AddUserExperienceWithTx(tx, userId, expGain, nowTime)
		if updateErr != nil {
			return updateErr
		}
		return nil
	})

	if err != nil {
		return false, 0, 0, metaerror.Wrap(err, "在事务中添加用户奖励经验值失败")
	}

	return false, newLevel, newExp, nil
}

// GetUserExperiences 获取用户的经验记录列表
func (d *UserDao) GetUserExperiences(ctx context.Context, userId int, limit int) ([]*foundationmodel.UserExperience, error) {
	var experiences []*foundationmodel.UserExperience
	db := d.db.WithContext(ctx).Model(&foundationmodel.UserExperience{}).
		Where("user_id = ?", userId).
		Order("inserter_time DESC")

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Find(&experiences).Error; err != nil {
		return nil, metaerror.Wrap(err, "get user experiences")
	}
	return experiences, nil
}

// GetUserExperienceTotal 获取用户的总经验值变化
func (d *UserDao) GetUserExperienceTotal(ctx context.Context, userId int) (int, error) {
	var total int
	err := d.db.WithContext(ctx).Model(&foundationmodel.UserExperience{}).
		Select("COALESCE(SUM(value), 0) as total").
		Where("user_id = ?", userId).
		Scan(&total).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "get user experience total")
	}
	return total, nil
}

// GetCheckinCount 获取指定日期的签到人数
func (d *UserDao) GetCheckinCount(ctx context.Context, date string) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&foundationmodel.UserExperience{}).
		Where("type = ? AND param = ?", foundationuser.ExperienceTypeCheckIn, date).
		Count(&count).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "查询签到人数失败")
	}
	return int(count), nil
}

// IsUserCheckedIn 检查用户在指定日期是否已签到
func (d *UserDao) IsUserCheckedIn(ctx context.Context, userId int, date string) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&foundationmodel.UserExperience{}).
		Where("user_id = ? AND type = ? AND param = ?", userId, foundationuser.ExperienceTypeCheckIn, date).
		Count(&count).Error
	if err != nil {
		return false, metaerror.Wrap(err, "查询用户签到状态失败")
	}
	return count > 0, nil
}

// AddUserCheckInCount 增加用户签到次数并添加经验值和金币（带防重复签到检查）
func (d *UserDao) AddUserCheckInCount(ctx context.Context, userId int, checkInCount int, expGain int, coinGain int, param string, nowTime time.Time) (bool, error) {
	// 检查用户是否已经签到过（基于唯一约束）
	exists, err := d.CheckUserExperienceExists(ctx, userId, foundationuser.ExperienceTypeCheckIn, param)
	if err != nil {
		return false, metaerror.Wrap(err, "检查用户签到记录失败")
	}
	if exists {
		return true, nil
	}

	hasDuplicate := false

	err = d.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 插入经验记录
		expRecord := foundationmodel.NewUserExperienceBuilder().
			UserId(userId).
			Value(expGain).
			Type(foundationuser.ExperienceTypeCheckIn).
			Param(param).
			InserterTime(nowTime).
			Build()
		insertErr := d.InsertUserExperienceWithTx(tx, expRecord)
		if insertErr != nil {
			// 判断是否是重复插入错误
			if errors.Is(insertErr, gorm.ErrDuplicatedKey) {
				hasDuplicate = true
				return nil
			}
			return insertErr
		}

		// 插入金币记录
		coinRecord := foundationmodel.NewUserCoinBuilder().
			UserId(userId).
			Value(coinGain).
			Type(foundationuser.CoinTypeCheckIn).
			Param(param).
			InserterTime(nowTime).
			Build()
		insertErr = d.InsertUserCoinWithTx(tx, coinRecord)
		if insertErr != nil {
			// 判断是否是重复插入错误
			if errors.Is(insertErr, gorm.ErrDuplicatedKey) {
				hasDuplicate = true
				return nil
			}
			return insertErr
		}

		// 更新用户签到次数
		res := tx.Model(&foundationmodel.User{}).Where("id = ?", userId).
			Updates(map[string]interface{}{
				"check_in_count": gorm.Expr("check_in_count + ?", checkInCount),
			})
		if res.Error != nil {
			return metaerror.Wrap(res.Error, "更新用户签到次数失败")
		}
		if res.RowsAffected == 0 {
			return metaerror.New("用户不存在")
		}

		// 更新用户经验值
		_, _, insertErr = d.AddUserExperienceWithTx(tx, userId, expGain, nowTime)
		if insertErr != nil {
			return insertErr
		}

		// 更新用户金币
		_, insertErr = d.AddUserCoinWithTx(tx, userId, coinGain, nowTime)
		if insertErr != nil {
			return insertErr
		}
		return nil
	})

	return hasDuplicate, err
}
