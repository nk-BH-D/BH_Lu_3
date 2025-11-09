package Lu

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	//"time"

	//orch "girhub.com/nk-BH-D/BH_Lu_3/bak/Orchestrator"
	calc "github.com/nk-BH-D/BH_Lu_3/bak/pkg"
)

type TaskInLu struct {
	TIL_ID    string `json:"til_id"`
	TIL_Value string `json:"til_value"`
}

// обрабатываем запрос от оркестратора
func LCF(w http.ResponseWriter, r *http.Request) {
	log.Println("LCF: начал обработку подзадачи")
	// проверка правильности метода
	if r.Method != http.MethodPost {
		log.Printf("метод: %s не поддерживаеться для этого запроса", r.Method)
		http.Error(w, "метод не разрешон", http.StatusMethodNotAllowed)
		return
	}

	// декодируем JSON
	var req_TIL TaskInLu
	log.Println("LCF: декодирует JSON")
	if err := json.NewDecoder(r.Body).Decode(&req_TIL); err != nil {
		log.Printf("ошибка JSON: %v", err)
		http.Error(w, "ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}
	log.Printf("LCF: декодировал JSON успешно: %+v", req_TIL)

	resultCalc, err := calc.Calc(req_TIL.TIL_Value)
	if err != nil {
		log.Printf("ошибка вычисления: %v", err)
		http.Error(w, "ошибка вычисления подзадачи", http.StatusInternalServerError)
		return
	}
	log.Printf("LCF: результат вычислений для %s: %s", req_TIL.TIL_ID, resultCalc)
	// создаём тело ответа оркестратору
	LCF_Body := struct {
		InLCF_ID     string `json:"in_lcf_id"`
		InLCF_Result string `json:"in_lcf_result"`
	}{
		InLCF_ID:     req_TIL.TIL_ID,
		InLCF_Result: resultCalc,
	}
	// кодируем JSON
	ilcfJSON, err := json.Marshal(LCF_Body)
	if err != nil {
		log.Printf("ошибка кодирования JSON: %v", err)
		http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
		return
	}
	// отправляем ответ оркестратору
	otvetResp, err := http.Post("http://orchestrator:8080/inLCF", "application/json", bytes.NewBuffer(ilcfJSON))
	if err != nil {
		log.Printf("ошибка при отправке ответа: %v", err)
		http.Error(w, "ошибка отправки ответа", http.StatusInternalServerError)
		return
	}
	defer otvetResp.Body.Close()

	//логируем ответ
	log.Printf("LCF: статус ответа: %s", otvetResp.Status)
}
