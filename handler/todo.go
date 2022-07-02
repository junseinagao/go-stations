package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/TechBowl-japan/go-stations/model"
	"github.com/TechBowl-japan/go-stations/service"
)

// A TODOHandler implements handling REST endpoints.
type TODOHandler struct {
	svc *service.TODOService
}

// NewTODOHandler returns TODOHandler based http.Handler.
func NewTODOHandler(svc *service.TODOService) *TODOHandler {
	return &TODOHandler{
		svc: svc,
	}
}

func (h *TODOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		prevId, err := strconv.Atoi(r.URL.Query().Get("prev_id"))
		if r.URL.Query().Get("prev_id") != "" && err != nil {
			log.Println(err)
			return
		} else if err != nil {
			prevId = 0
		}
		size, err := strconv.Atoi(r.URL.Query().Get("size"))
		if r.URL.Query().Get("size") != "" && err != nil {
			log.Println(err)
			return
		} else if err != nil {
			size = 5
		}
		response, err := h.Read(r.Context(), &model.ReadTODORequest{
			PrevID: int64(prevId),
			Size:   int64(size),
		})
		if err != nil {
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	case http.MethodPost:
		// * request body を decode する
		var requestBody model.CreateTODORequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Println(err)
			return
		}
		// * subject が 空文字の場合は 早期return
		if requestBody.Subject == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// * DB に TODO を保存
		response, err := h.Create(r.Context(), &requestBody)
		if err != nil {
			log.Println(err)
			return
		}

		// * HTTP Responseを返す
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return

	case http.MethodPut:
		// * request body を decode する
		var requestBody model.UpdateTODORequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// * ID が 空文字の場合 または Subject が 空文字の場合は 早期return
		if requestBody.Subject == "" || requestBody.ID == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// * DB の TODO を更新
		response, err := h.Update(r.Context(), &requestBody)
		if err != nil {
			log.Println(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// * HTTP Responseを返す
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	case http.MethodDelete:
		// * request body を decode する
		var requestBody model.DeleteTODORequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Println(err)
			return
		}
		if len(requestBody.IDs) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// * DB の TODO を削除
		response, err := h.Delete(r.Context(), &requestBody)
		var errNotFound *model.ErrNotFound
		if errors.As(err, &errNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	default:
		return
	}
}

// Create handles the endpoint that creates the TODO.
func (h *TODOHandler) Create(ctx context.Context, req *model.CreateTODORequest) (*model.CreateTODOResponse, error) {
	todo, err := h.svc.CreateTODO(ctx, req.Subject, req.Description)
	return &model.CreateTODOResponse{
		TODO: *todo,
	}, err
}

// Read handles the endpoint that reads the TODOs.
func (h *TODOHandler) Read(ctx context.Context, req *model.ReadTODORequest) (*model.ReadTODOResponse, error) {
	todoPointers, err := h.svc.ReadTODO(ctx, req.PrevID, req.Size)

	// TODO スライスではなく配列にする (スライスではなく配列で生成したかったが、len(todoPointers) は誤り)

	todos := make([]model.TODO, 0)
	for _, todo := range todoPointers {
		todos = append(todos, *todo)
	}
	return &model.ReadTODOResponse{
		TODOs: todos,
	}, err
}

// Update handles the endpoint that updates the TODO.
func (h *TODOHandler) Update(ctx context.Context, req *model.UpdateTODORequest) (*model.UpdateTODOResponse, error) {
	todo, err := h.svc.UpdateTODO(ctx, req.ID, req.Subject, req.Description)
	return &model.UpdateTODOResponse{
		TODO: *todo,
	}, err
}

// Delete handles the endpoint that deletes the TODOs.
func (h *TODOHandler) Delete(ctx context.Context, req *model.DeleteTODORequest) (*model.DeleteTODOResponse, error) {
	err := h.svc.DeleteTODO(ctx, req.IDs)
	return &model.DeleteTODOResponse{}, err
}
