package server

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func (s *Server) ListNotes() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		notes, err := s.NH.List(ctx)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		notesJson, err := json.Marshal(notes)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		writer.Write(notesJson)
	}
}

func (s *Server) AddNote() http.HandlerFunc {
	type requestBody struct {
		Text       string `json:"text"`
		Expiration int    `json:"expiration"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		var r requestBody
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(bytes, &r)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		ctx := request.Context()
		uid, err := uuid.NewV4()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		note, err := s.NH.Create(ctx, uid, r.Text, r.Expiration)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		noteJson, err := json.Marshal(note)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		writer.Write(noteJson)
	}
}

func (s *Server) UpdateNote() http.HandlerFunc {
	type requestBody struct {
		Text string `json:"text"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		var r requestBody
		ctx := request.Context()
		vars := mux.Vars(request)
		uidStr := vars["uid"]
		uid, err := uuid.FromString(uidStr)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(bytes, &r)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		note, err := s.NH.Update(ctx, uid, r.Text)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		noteJson, err := json.Marshal(note)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		writer.Write(noteJson)
	}
}

func (s *Server) GetNote() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		uidStr := vars["uid"]
		uid, err := uuid.FromString(uidStr)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := request.Context()

		note, err := s.NH.Get(ctx, uid)
		if err != nil {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		nJson, err := json.Marshal(note)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(nJson)
	}
}

func (s *Server) DeleteNote() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		uidStr := vars["uid"]
		uid, err := uuid.FromString(uidStr)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := request.Context()

		_, err = s.NH.Delete(ctx, uid)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) PeekNote() http.HandlerFunc {
	type responseBody struct {
		Exist bool `json:"exist"`
	}
	type requestBody struct {
		ID uuid.UUID `json:"id"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		var r requestBody
		ctx := request.Context()
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(bytes, &r)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		note, err := s.NH.Get(ctx, r.ID)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if note == nil {
			response := responseBody{Exist: false}
			jsonResp, err := json.Marshal(response)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			writer.WriteHeader(http.StatusNotFound)
			writer.Write(jsonResp)
			return
		}

		response := responseBody{Exist: true}
		jsonResp, err := json.Marshal(response)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(jsonResp)
	}
}

func (s *Server) PopNote() http.HandlerFunc {
	type requestBody struct {
		ID uuid.UUID `json:"id"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		var r requestBody
		ctx := request.Context()
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(bytes, &r)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		note, err := s.NH.Delete(ctx, r.ID)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if note == nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		noteJson, err := json.Marshal(note)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(noteJson)
	}
}
