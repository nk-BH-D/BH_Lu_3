package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	id_uuid "github.com/google/uuid"
	tokens_BH "github.com/nk-BH-D/BH_Lu_3/bak/token"
)

// структура для представления выражения
type Expression_BH struct {
	ID         string     `json:"id"`
	Expression string     `json:"expr"`
	Status     string     `json:"status"`
	Result     string     `json:"result"`
	CreatedAt  time.Time  `json:"created_at"`
	Tasks      []*Task_BH `json:"task_BH"`
}

// структура для представления под выражения
type Task_BH struct {
	Task_ID       string    `json:"task_id"`
	Expression_ID string    `json:"expr_id"`
	Arg_1         string    `json:"arg_1"`
	Operator      string    `json:"operator"`
	Arg_2         string    `json:"arg_2"`
	Task_Status   string    `json:"task_status"`
	Task_Result   string    `json:"task_result"`
	CreatedAtTask time.Time `json:"created_at_task"`
}

// хранение актуальных данных в памяти
type MemoryData struct {
	Expression map[string]*Expression_BH
	Task       map[string]*Task_BH
	mu         sync.RWMutex
	TTL        time.Duration
}

// функция для очистки памяти TTL(так то всё можно писать в базу данных но кеш актуальной информации оптимизирует работу программы)
func (md *MemoryData) StartCleanup() {
	//горутина для удаления подзадач
	go func() {
		for {
			time.Sleep(30 * time.Second)
			// читаем с RMutex
			md.mu.RLock()
			expiredExpressions := []string{}
			for expr_id, expr := range md.Expression {
				if time.Since(expr.CreatedAt) > md.TTL {
					expiredExpressions = append(expiredExpressions, expr_id)
				}
			}
			md.mu.RUnlock()
			// удаляем с Mutex
			md.mu.Lock()
			for _, expr_id := range expiredExpressions {
				log.Printf("Удаляем задачу с ID: %s", expr_id)
				delete(md.Expression, expr_id)
			}
			md.mu.Unlock()
		}
	}()
	// горутина для удаления подзадач
	go func() {
		for {
			time.Sleep(30 * time.Second)
			// читаем с RMutex
			md.mu.RLock()
			expiredTask := []string{}
			for task_id, task := range md.Task {
				if time.Since(task.CreatedAtTask) > md.TTL {
					expiredTask = append(expiredTask, task_id)
				}
			}
			md.mu.RUnlock()
			// удаляем с Mutex
			md.mu.Lock()
			for _, task_id := range expiredTask {
				log.Printf("Удаляем подзадачу с ID: %s", task_id)
				delete(md.Task, task_id)
			}
			md.mu.Unlock()
		}
	}()
}

// инициализируем мапы
func NewMemoryData(ttl time.Duration) *MemoryData {
	md := &MemoryData{
		Expression: make(map[string]*Expression_BH),
		Task:       make(map[string]*Task_BH),
		TTL:        ttl,
	}
	md.StartCleanup()
	return md
}

// дабавления выражения с которым работает программа в память
func AddActualExpression(md *MemoryData, expr *Expression_BH) {
	md.mu.Lock()
	defer md.mu.Unlock()
	log.Printf("AddActualExpression: кешировал: ID: %s", expr.ID)
	md.Expression[expr.ID] = expr
}

func AddActualTask(md *MemoryData, task *Task_BH) {
	md.mu.Lock()
	defer md.mu.Unlock()
	log.Printf("AddActualTask: кешировал: ID: %s", task.Task_ID)
	md.Task[task.Task_ID] = task
}

func GenerateID() (string, error) {
	id, err := id_uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации ID: %w", err)
	}
	return id.String(), nil
}

func CalculateHandler(w http.ResponseWriter, r *http.Request, md *MemoryData) {
	log.Println("CH: начал обработку выражения")
	//проверяем правильный ли метод
	if r.Method != http.MethodPost {
		log.Printf("метод: %s не поддерживается для этого запроса", r.Method)
		http.Error(w, "метод не разрешон", http.StatusMethodNotAllowed)
		return
	}
	// декодируем JSON
	var req_EBH Expression_BH
	log.Println("CH: декодирует JSON")
	if err := json.NewDecoder(r.Body).Decode(&req_EBH); err != nil {
		log.Printf("ошибка JSON: %v", err)
		http.Error(w, "ошибки при декодировании JSON", http.StatusBadRequest)
		return
	}
	log.Printf("CH: декодировал JSON успешно: %+v", req_EBH)
	// генерируем ID
	id, err := GenerateID()
	log.Println("CH: генерирует ID")
	if err != nil {
		log.Printf("ошибка generateID: %v", err)
		http.Error(w, "ошибка генерации ID", http.StatusInternalServerError)
		return
	}
	log.Printf("CH: сгенерировал ID успешно: %s", id)
	//создаём указатель на выражение что бы работать с ним внутри функции
	expr := &Expression_BH{
		ID:         id,
		Expression: req_EBH.Expression,
		Status:     "созданно",
		Result:     "",
		CreatedAt:  time.Now(),
		Tasks:      []*Task_BH{},
	}
	log.Printf("CH: создал выражение:\nID: %s\nExpression: %s\nStatus: %s\nResult: %s\nCreatedAt: %v", expr.ID, expr.Expression, expr.Status, expr.Result, expr.CreatedAt)
	// записываем данные из expr в память(позже запишем всё выражение в db)
	md.mu.Lock()
	md.Expression[id] = expr
	md.mu.Unlock()
	AddActualExpression(md, expr) // сам процес записи в память
	// токенизируем данные что бы в дальнейшем с ними было удобней работать\
	tokens, err := tokens_BH.Tokenize_BH(expr.Expression)
	if err != nil {
		expr.Status = "error"
		log.Printf("ошибка токенизации: %v", err)
		http.Error(w, "ошибка токенизации", http.StatusBadRequest)
		return
	}
	log.Printf("CH: токены: %+v", tokens)
	// предаём данные в главу каскада решателя
	result, err := Calcualtor(tokens, expr, md)
	if err != nil {
		expr.Status = "error"
		log.Printf("ошибка в процессе решения: %v", err)
		http.Error(w, "ошибка в процессе решения", http.StatusInternalServerError)
		return
	}
	log.Printf("РЕЗУЛЬТАТ: %s", result)
	// отправляем результат пользователю
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"result": result,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка при отправке результата: %v", err)
		http.Error(w, "ошибка при отправке результата", http.StatusInternalServerError)
		return
	}
}

// глава каскада
func Calcualtor(tokens []tokens_BH.Token, expr *Expression_BH, md *MemoryData) (string, error) {
	task_result_slice, err := OpenParent(tokens, expr, md)
	if err != nil {
		return "", fmt.Errorf("ошибка раскрытия скобок: %w", err)
	}
	log.Printf("Calculator: получил слайс с результатами из OP: %+v", task_result_slice)
	// заменяем результаты на место скобок
	expression := expr.Expression
	for _, result := range task_result_slice {
		// находим первую пару скобок
		openIndex := -1
		closeIndex := -1
		openCount := 0
		for i, char := range expression {
			if char == '(' {
				if openCount == 0 {
					openIndex = i
				}
				openCount++
			} else if char == ')' {
				openCount--
				if openCount == 0 {
					closeIndex = i
					break
				}
			}
		}
		// если не нашли скобки, выходим
		if openIndex == -1 || closeIndex == -1 {
			break
		}
		// заменяем подвыражение на результат
		subExpression := expression[openIndex : closeIndex+1]
		expression = expression[:openIndex] + result + expression[closeIndex+1:]
		log.Printf("Calculator: заменил подвыражение %s на результат %s", subExpression, result)
		expr.Expression = expression
		log.Printf("Calculator: обновлённое выражение после OP: %s", expr.Expression)
	}
	// Обновляем выражение в структуре

	// отправляем обновлённое выражение без скобок дальше по каскаду
	//openTheDoor := expr.Expression
	tokensOTD, err := tokens_BH.Tokenize_BH(expression)
	if err != nil {
		return "", fmt.Errorf("ошибка повторной токенизации: %w", err)
	}
	afteMAD_Slice, err := MultiplicationAndDivision(tokensOTD, expr, md)
	if err != nil {
		return "", fmt.Errorf("ошибка умножения и деления: %w", err)
	}
	log.Printf("Calculator: получил слайс с результатами из MAD: %+v", afteMAD_Slice)
	// вставляем результаты на место задач
	expression = expr.Expression
	for _, resultMAD := range afteMAD_Slice {
		for i := 0; i < len(tokensOTD); i++ {
			if tokensOTD[i].Type == tokens_BH.TOKEN_OPERATOR && tokensOTD[i].Value == "*" || tokensOTD[i].Type == tokens_BH.TOKEN_OPERATOR && tokensOTD[i].Value == "/" {
				// получаем аргументы и оператор
				arg1 := tokensOTD[i-1].Value
				operatorr := tokensOTD[i].Value
				arg2 := tokensOTD[i+1].Value
				// формируем подвыражение
				subExpression := arg1 + operatorr + arg2
				// заменяем подвыражение на результат
				expression = strings.ReplaceAll(expression, subExpression, resultMAD)
				log.Printf("Calculator: заменил подвыражение %s на результат %s", subExpression, resultMAD)
				// обновляем выражение в структуре
				expr.Expression = expression
				log.Printf("Calculator: обновлённое выражение после MAD выражение: %s", expr.Expression)
				// обновляем токены
				tokensOTD = append(tokensOTD[:i-1], tokensOTD[i+2:]...)
				break
			}
		}
	}

	//MAD := expr.Expression

	tokensMAD, err := tokens_BH.Tokenize_BH(expression)
	if err != nil {
		return "", fmt.Errorf("ошибка повторной токенизации: %w", err)
	}

	afterAAS_slice, err := AdditionAndSubtraction(tokensMAD, expr, md)
	if err != nil {
		return "", fmt.Errorf("ошибка сложения и вычитания")
	}
	log.Printf("Calculator: получил результаты из AAS: %s", afterAAS_slice)
	// вставляем результат на место подзадач
	expression = expr.Expression
	for _, resultASS := range afterAAS_slice {
		for i := 0; i < len(tokensMAD); i++ {
			if tokensMAD[i].Type == tokens_BH.TOKEN_OPERATOR && tokensMAD[i].Value == "+" || tokensMAD[i].Type == tokens_BH.TOKEN_OPERATOR && tokensMAD[i].Value == "-" {
				arg1 := tokensMAD[i-1].Value
				operatorr := tokensMAD[i].Value
				arg2 := tokensMAD[i+1].Value
				// формируем подвыражения
				subExpression := arg1 + operatorr + arg2
				//заменяем подвыражения на результата
				expression = strings.ReplaceAll(expression, subExpression, resultASS)
				log.Printf("Calculator: заменил подвыражение %s на результат %s", subExpression, resultASS)
				expr.Expression = expression
				log.Printf("Calculator: обновил выражение после AAS: %s", expr.Expression)
				// обновляем токены
				tokensMAD = append(tokensMAD[:i-1], tokensMAD[i+2:]...)
				break
			}
		}
	}
	return expr.Expression, nil
}

func OpenParent(tokens []tokens_BH.Token, expr *Expression_BH, md *MemoryData) ([]string, error) {
	var (
		allOpenParentSlice [][]tokens_BH.Token
	)
	task_result_slice := make([]string, 0)
	for i, openTheDoor := range tokens {
		if openTheDoor.Type == tokens_BH.TOKEN_PARENT_OPEN {
			closeIndex := -1
			openIndex := 1
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j].Type == tokens_BH.TOKEN_PARENT_OPEN {
					openIndex++
				} else if tokens[j].Type == tokens_BH.TOKEN_PARENT_CLOSE {
					openIndex--
					if openIndex == 0 {
						closeIndex = j
						break
					}
				}
			}
			if closeIndex == -1 {
				return nil, fmt.Errorf("не найдена закрывающая скобка")
			}
			allOpenParentSlice = append(allOpenParentSlice, tokens[i+1:closeIndex])
			log.Printf("OP: openParentSlice: %+v", allOpenParentSlice)
		}
	}
	// цикл в котором присваиваем переменные значения и создаём задачи
	for _, openParentSlice := range allOpenParentSlice {
		// генерируем ID для Task
		task_id, err := GenerateID()
		if err != nil {
			return nil, fmt.Errorf("ошибка генерации ID для Task: %v", err)
		}
		// индекс который показывает какое число на каком месте стоит
		numIndex := 0
		// объявляем переменные вне цикла что бы попасть в зону видимости
		var (
			arg_1    string
			operator string
			arg_2    string
		)
		for _, taskConstruct := range openParentSlice {
			if taskConstruct.Type == tokens_BH.TOKEN_NUMBER && numIndex == 0 {
				arg_1 = taskConstruct.Value
				numIndex++
			} else if taskConstruct.Type == tokens_BH.TOKEN_OPERATOR {
				operator = taskConstruct.Value
			} else if taskConstruct.Type == tokens_BH.TOKEN_NUMBER && numIndex == 1 {
				arg_2 = taskConstruct.Value
			}
		}
		// создаём подзадачу
		task := &Task_BH{
			Task_ID:       task_id,
			Expression_ID: expr.ID,
			Arg_1:         arg_1,
			Operator:      operator,
			Arg_2:         arg_2,
			Task_Status:   "созданно",
			Task_Result:   "",
			CreatedAtTask: time.Now(),
		}
		log.Printf("OP: создалa подзадачу:\nTask_ID: %s\nExpression_ID: %s\nArg_1: %s\nOperator: %s\nArg_2: %s\nTask_Status: %s\nTask_Result: %s\nCreatedAtTask: %v", task.Task_ID, task.Expression_ID, task.Arg_1, task.Operator, task.Arg_2, task.Task_Status, task.Task_Result, task.CreatedAtTask)
		// запишем подзадачу в Task
		expr.Tasks = append(expr.Tasks, task)
		// записываем созданные подзадачи в память
		md.mu.Lock()
		md.Task[task_id] = task
		md.mu.Unlock()
		AddActualTask(md, task)
		// передаём данные в функцию которая отправит их агнету
		err = SendTaskAgent(task)
		if err != nil {
			return nil, fmt.Errorf("ошибка отправки запроса агенту: %w", err)
		}
		md.mu.Lock()
		task_result := md.Task[task_id].Task_Result
		md.mu.Unlock()
		task_result_slice = append(task_result_slice, task_result)
	}
	return task_result_slice, nil
}

// фуекция для обработки умножения и деления
func MultiplicationAndDivision(tokensOTD []tokens_BH.Token, expr *Expression_BH, md *MemoryData) ([]string, error) {
	log.Printf("MAD: приняла значение: %+v", tokensOTD)
	MAD_result_slice := make([]string, 0)
	for i, token := range tokensOTD {
		if token.Type == tokens_BH.TOKEN_OPERATOR && token.Value == "*" || token.Type == tokens_BH.TOKEN_OPERATOR && token.Value == "/" {
			var (
				arg_1    string
				operator string
				arg_2    string
			)
			arg_1 = tokensOTD[i-1].Value
			operator = token.Value
			arg_2 = tokensOTD[i+1].Value
			task_id, err := GenerateID()
			if err != nil {
				return nil, fmt.Errorf("ошибка генерации ID: %w", err)
			}
			// создаём подзадачу
			task := &Task_BH{
				Task_ID:       task_id,
				Expression_ID: expr.ID,
				Arg_1:         arg_1,
				Operator:      operator,
				Arg_2:         arg_2,
				Task_Status:   "созданно",
				Task_Result:   "",
				CreatedAtTask: time.Now(),
			}
			log.Printf("MAD: создалa подзадачу:\nTask_ID: %s\nExpression_ID: %s\nArg_1: %s\nOperator: %s\nArg_2: %s\nTask_Status: %s\nTask_Result: %s\nCreatedAtTask: %v", task.Task_ID, task.Expression_ID, task.Arg_1, task.Operator, task.Arg_2, task.Task_Status, task.Task_Result, task.CreatedAtTask)
			// запишем подзадачу в Task
			expr.Tasks = append(expr.Tasks, task)
			// записываем созданные подзадачи в память
			md.mu.Lock()
			md.Task[task_id] = task
			md.mu.Unlock()
			AddActualTask(md, task)
			// отправляем запрос агенту
			err = SendTaskAgent(task)
			if err != nil {
				return nil, fmt.Errorf("ошибка отправки запроса агенту: %w", err)
			}
			md.mu.Lock()
			MAD_result := md.Task[task_id].Task_Result
			md.mu.Unlock()
			MAD_result_slice = append(MAD_result_slice, MAD_result)
		}
	}
	return MAD_result_slice, nil
}

// функция для обработки сложения и вычитания
func AdditionAndSubtraction(tokensMAD []tokens_BH.Token, expr *Expression_BH, md *MemoryData) ([]string, error) {
	log.Printf("AAS: получил токены: %+v", tokensMAD)
	AAS_result_slice := make([]string, 0)
	for i, token := range tokensMAD {
		if token.Type == tokens_BH.TOKEN_OPERATOR && token.Value == "+" || token.Type == tokens_BH.TOKEN_OPERATOR && token.Value == "-" {
			var (
				arg_1    string
				operator string
				arg_2    string
			)
			arg_1 = tokensMAD[i-1].Value
			operator = token.Value
			arg_2 = tokensMAD[i+1].Value
			task_id, err := GenerateID()
			if err != nil {
				return nil, fmt.Errorf("ошибка генериции ID: %w", err)
			}
			// создаём подзадачу
			task := &Task_BH{
				Task_ID:       task_id,
				Expression_ID: expr.ID,
				Arg_1:         arg_1,
				Operator:      operator,
				Arg_2:         arg_2,
				Task_Status:   "созданно",
				Task_Result:   "",
				CreatedAtTask: time.Now(),
			}
			log.Printf("MAD: создалa подзадачу:\nTask_ID: %s\nExpression_ID: %s\nArg_1: %s\nOperator: %s\nArg_2: %s\nTask_Status: %s\nTask_Result: %s\nCreatedAtTask: %v", task.Task_ID, task.Expression_ID, task.Arg_1, task.Operator, task.Arg_2, task.Task_Status, task.Task_Result, task.CreatedAtTask)
			// запишем подзадачу в Task
			expr.Tasks = append(expr.Tasks, task)
			// записываем созданные подзадачи в память
			md.mu.Lock()
			md.Task[task_id] = task
			md.mu.Unlock()
			AddActualTask(md, task)
			//отправляем запрос агенту
			err = SendTaskAgent(task)
			if err != nil {
				return nil, fmt.Errorf("ошибка отправки запроса агенту: %w", err)
			}
			md.mu.Lock()
			AAS_result := md.Task[task_id].Task_Result
			md.mu.Unlock()
			AAS_result_slice = append(AAS_result_slice, AAS_result)
		}
	}
	return AAS_result_slice, nil
}

// отправляем запрос агенту
func SendTaskAgent(task *Task_BH) error {
	body := struct {
		TIL_ID    string `json:"til_id"`
		TIL_Value string `json:"til_value"`
	}{
		TIL_ID:    task.Task_ID,
		TIL_Value: fmt.Sprintf("%s%s%s", task.Arg_1, task.Operator, task.Arg_2),
	}
	// кодируем строку в JSON
	JSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ошибка кодирования JSON: %v", err)
	}
	// Отправляем POST-запрос
	resp, err := http.Post("http://demon:8081/LCF", "application/json", bytes.NewBuffer(JSON))
	if err != nil {
		log.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Логируем статус ответа
	log.Printf("Статус ответа: %s", resp.Status)
	return nil
}

// вспомогательня структура для размещения данных
type LCFOS struct {
	LCFOS_ID     string `json:"in_lcf_id"`
	LCFOS_Result string `json:"in_lcf_result"`
}

// получает ответ от агента
func LCF_Otvet(w http.ResponseWriter, r *http.Request, md *MemoryData) {
	log.Println("LCFO: начал обработку")

	if r.Method != http.MethodPost {
		log.Printf("метод: %s не поддерживаеться для этого запроса", r.Method)
		http.Error(w, "метод не разрешон", http.StatusMethodNotAllowed)
		return
	}

	var req_LCFOS LCFOS
	log.Println("LCFO: декодирует JSON")
	if err := json.NewDecoder(r.Body).Decode(&req_LCFOS); err != nil {
		log.Printf("ошибка JSON: %v", err)
		http.Error(w, "ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}
	log.Printf("LCFO: декодировал JSON успешно: %+v", req_LCFOS)
	//обновим данные
	newTask := &Task_BH{
		Task_ID:       req_LCFOS.LCFOS_ID,
		Task_Status:   "обработанно",
		Task_Result:   req_LCFOS.LCFOS_Result,
		CreatedAtTask: time.Now(),
	}
	md.mu.Lock()
	md.Task[req_LCFOS.LCFOS_ID] = newTask
	md.mu.Unlock()
	AddActualTask(md, newTask)
	log.Printf("OP: обработынный Task:\nTask_ID: %s\nTask_Status: %s\nTask_Result %s", newTask.Task_ID, newTask.Task_Status, newTask.Task_Result)
}
