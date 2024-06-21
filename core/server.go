package core

// var requestID = 0
// var queue = NewMessageQueue()

// func main() {
// 	// Start workers
// 	for i := 1; i <= 2; i++ {
// 		go Worker(i, queue)
// 	}

// 	http.HandleFunc("/process", ProcessRequestHandler)
// 	fmt.Println("Server started at :8080")
// 	http.ListenAndServe(":8080", nil)
// }

// func ProcessRequestHandler(w http.ResponseWriter, r *http.Request) {
// 	requestID++

// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	// Encode the request body as binary
// 	var encodedMessage bytes.Buffer
// 	err = binary.Write(&encodedMessage, binary.LittleEndian, body)
// 	if err != nil {
// 		http.Error(w, "Failed to encode request body", http.StatusInternalServerError)
// 		return
// 	}

// 	responseChan := make(chan []byte)
// 	req := Request{
// 		ID:           requestID,
// 		Message:      encodedMessage.Bytes(),
// 		ResponseChan: responseChan,
// 	}

// 	queue.Enqueue(req)

// 	response := <-responseChan

// 	w.Header().Set("Content-Type", "application/octet-stream")
// 	w.Write(response)
// }
