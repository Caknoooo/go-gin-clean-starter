package controller

import (
	"net/http"

	"github.com/Caknoooo/golang-clean_template/dto"
	"github.com/Caknoooo/golang-clean_template/entities"
	"github.com/Caknoooo/golang-clean_template/services"
	"github.com/Caknoooo/golang-clean_template/utils"
	"github.com/gin-gonic/gin"
)

type UserController interface {
	RegisterUser(ctx *gin.Context)
	GetAllUser(ctx *gin.Context)
	MeUser(ctx *gin.Context)
	LoginUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)
	DeleteUser(ctx *gin.Context)
}

type userController struct {
	jwtService  services.JWTService
	userService services.UserService
}

func NewUserController(us services.UserService, jwt services.JWTService) UserController {
	return &userController{
		jwtService:  jwt,
		userService: us,
	}
}

func (uc *userController) RegisterUser(ctx *gin.Context) {
	var user dto.UserCreateDTO
	if err := ctx.ShouldBind(&user); err != nil {
		panic(err) // harus diperbaiki
	}

	if checkUser, _ := uc.userService.CheckUser(ctx.Request.Context(), user.Email); checkUser {
		res := utils.BuildResponseFailed("Email Sudah Terdaftar", "Failed", utils.EmptyObj{})
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	result, err := uc.userService.RegisterUser(ctx.Request.Context(), user)
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Menambahkan User", "Failed", utils.EmptyObj{})
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Berhasil Menambahkan User", result)
	ctx.JSON(http.StatusOK, res)
}

func (uc *userController) GetAllUser(ctx *gin.Context) {
	result, err := uc.userService.GetAllUser(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Mendapatkan List User", err.Error(), utils.EmptyObj{})
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess("Berhasil Mendapatkan List User", result)
	ctx.JSON(http.StatusOK, res)
}

func (uc *userController) MeUser(ctx *gin.Context) {
	token := ctx.MustGet("token").(string)
	userID, err := uc.jwtService.GetUserIDByToken(token)
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Memproses Request", "Token Tidak Valid", nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	result, err := uc.userService.GetUserByID(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Mendapatkan User", err.Error(), utils.EmptyObj{})
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Berhasil Mendapatkan User", result)
	ctx.JSON(http.StatusOK, res)
}

func (uc *userController) LoginUser(ctx *gin.Context) {
	var userLoginDTO dto.UserLoginDTO
	err := ctx.ShouldBind(&userLoginDTO)
	res, _ := uc.userService.Verify(ctx.Request.Context(), userLoginDTO.Email, userLoginDTO.Password)
	if !res {
		response := utils.BuildResponseFailed("Gagal Login", "Email atau Password Salah", utils.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := uc.userService.GetUserByEmail(ctx.Request.Context(), userLoginDTO.Email)
	if err != nil {
		response := utils.BuildResponseFailed("Gagal Login", err.Error(), utils.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	token := uc.jwtService.GenerateToken(user.ID, user.Role)
	userResponse := entities.Authorization{
		Token: token,
		Role:  user.Role,
	}

	response := utils.BuildResponseSuccess("Berhasil Login", userResponse)
	ctx.JSON(http.StatusOK, response)
}

func (uc *userController) UpdateUser(ctx *gin.Context) {
	var userDTO dto.UserUpdateDTO
	if err := ctx.ShouldBind(&userDTO); err != nil {
		res := utils.BuildResponseFailed("Gagal Mendapatkan Request Dari Body", err.Error(), utils.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	token := ctx.MustGet("token").(string)
	userID, err := uc.jwtService.GetUserIDByToken(token)
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Memproses Request", "Token Tidak Valid", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	userDTO.ID = userID
	if err = uc.userService.UpdateUser(ctx.Request.Context(), userDTO); err != nil {
		res := utils.BuildResponseFailed("Gagal Update User", err.Error(), utils.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Berhasil Update User", userDTO)
	ctx.JSON(http.StatusOK, res)
}

func (uc *userController) DeleteUser(ctx *gin.Context) {
	token := ctx.MustGet("token").(string)
	userID, err := uc.jwtService.GetUserIDByToken(token)
	if err != nil {
		res := utils.BuildResponseFailed("Gagal Memproses Request", "Token Tidak Valid", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err = uc.userService.DeleteUser(ctx.Request.Context(), userID); err != nil {
		res := utils.BuildResponseFailed("Gagal Menghapus User", err.Error(), utils.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Berhasil Menghapus User", utils.EmptyObj{})
	ctx.JSON(http.StatusOK, res)
}