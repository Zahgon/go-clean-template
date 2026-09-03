package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gin-gonic/gin"
)

// @Summary     Register
// @Description Register a new user
// @ID          register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Register true "Registration data"
// @Success     201     {object} entity.User
// @Failure     400     {object} response.Error
// @Failure     409     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/register [post]
func (r *V1) register(ctx *gin.Context) {
	var body request.Register

	if err := ctx.ShouldBindJSON(&body); err != nil {
		r.l.Error(err, "restapi - v1 - register")
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")

		return
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - register")
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")

		return
	}

	user, err := r.u.Register(ctx.Request.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		r.l.Error(err, "restapi - v1 - register")

		if errors.Is(err, entity.ErrUserAlreadyExists) {
			errorResponse(ctx, http.StatusConflict, "user already exists")

			return
		}

		errorResponse(ctx, http.StatusInternalServerError, "internal server error")

		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// @Summary     Login
// @Description Authenticate user and get JWT token
// @ID          login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body     request.Login true "Login credentials"
// @Success     200     {object} response.Token
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     500     {object} response.Error
// @Router      /auth/login [post]
func (r *V1) login(ctx *gin.Context) {
	var body request.Login

	if err := ctx.ShouldBindJSON(&body); err != nil {
		r.l.Error(err, "restapi - v1 - login")
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")

		return
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - login")
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")

		return
	}

	token, err := r.u.Login(ctx.Request.Context(), body.Email, body.Password)
	if err != nil {
		r.l.Error(err, "restapi - v1 - login")

		if errors.Is(err, entity.ErrInvalidCredentials) {
			errorResponse(ctx, http.StatusUnauthorized, "invalid credentials")

			return
		}

		errorResponse(ctx, http.StatusInternalServerError, "internal server error")

		return
	}

	ctx.JSON(http.StatusOK, response.Token{Token: token})
}

// @Summary     Get profile
// @Description Get current user profile
// @ID          profile
// @Tags        user
// @Produce     json
// @Success     200 {object} entity.User
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/profile [get]
func (r *V1) profile(ctx *gin.Context) {
	userID, ok := extractUserID(ctx)
	if !ok {
		errorResponse(ctx, http.StatusUnauthorized, "unauthorized")

		return
	}

	user, err := r.u.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - profile")

		if errors.Is(err, entity.ErrUserNotFound) {
			errorResponse(ctx, http.StatusNotFound, "user not found")

			return
		}

		errorResponse(ctx, http.StatusInternalServerError, "internal server error")

		return
	}

	ctx.JSON(http.StatusOK, user)
}
