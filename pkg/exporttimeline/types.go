package exporttimeline

type BucketRow struct {
	SessionID       string `json:"session_id"`
	BucketIdx       int    `json:"bucket_idx"`
	BucketStart     string `json:"bucket_start"`
	BucketEnd       string `json:"bucket_end"`
	TurnCount       int    `json:"turn_count"`
	ToolCallCount   int    `json:"tool_call_count"`
	ReadCount       int    `json:"read_count"`
	ModifyCount     int    `json:"modify_count"`
	NewCount        int    `json:"new_count"`
	ExecuteCount    int    `json:"execute_count"`
	ErrorCount      int    `json:"error_count"`
	ActiveFileCount int    `json:"active_file_count"`
	JumpTurnIdx     *int   `json:"jump_turn_idx"`
}

type FileActivityRow struct {
	SessionID         string `json:"session_id"`
	FilePath          string `json:"file_path"`
	FileRank          int    `json:"file_rank"`
	TotalOperations   int    `json:"total_operations"`
	BucketIdx         int    `json:"bucket_idx"`
	Operations        int    `json:"operations"`
	ReadCount         int    `json:"read_count"`
	ModifyCount       int    `json:"modify_count"`
	NewCount          int    `json:"new_count"`
	ExecuteCount      int    `json:"execute_count"`
	DominantOperation string `json:"dominant_operation"`
}

type IdleWindowRow struct {
	SessionID             string `json:"session_id"`
	StartBucketIdx        int    `json:"start_bucket_idx"`
	EndBucketIdx          int    `json:"end_bucket_idx"`
	IdleStart             string `json:"idle_start"`
	IdleEnd               string `json:"idle_end"`
	BucketCount           int    `json:"bucket_count"`
	ToolCallCountInWindow int    `json:"tool_call_count_in_window"`
}

type PhaseSignalRow struct {
	SessionID                 string `json:"session_id"`
	BucketIdx                 int    `json:"bucket_idx"`
	ToolCallCount             int    `json:"tool_call_count"`
	ReadCount                 int    `json:"read_count"`
	ModifyCount               int    `json:"modify_count"`
	NewCount                  int    `json:"new_count"`
	ExecuteCount              int    `json:"execute_count"`
	ErrorCount                int    `json:"error_count"`
	ValidationSignalCount     int    `json:"validation_signal_count"`
	CodeReadSignalCount       int    `json:"code_read_signal_count"`
	ImplementationSignalCount int    `json:"implementation_signal_count"`
	DominantPhaseSignal       string `json:"dominant_phase_signal"`
}

type ThreadSignalRow struct {
	SessionID       string `json:"session_id"`
	FilePath        string `json:"file_path"`
	StartBucketIdx  int    `json:"start_bucket_idx"`
	EndBucketIdx    int    `json:"end_bucket_idx"`
	BucketSpan      int    `json:"bucket_span"`
	TotalOperations int    `json:"total_operations"`
	JumpTurnIdx     *int   `json:"jump_turn_idx"`
}

type SQLTimelineData struct {
	Buckets       []BucketRow       `json:"buckets"`
	FileActivity  []FileActivityRow `json:"file_activity"`
	IdleWindows   []IdleWindowRow   `json:"idle_windows"`
	PhaseSignals  []PhaseSignalRow  `json:"phase_signals"`
	ThreadSignals []ThreadSignalRow `json:"thread_signals"`
}

type LoadOptions struct {
	QueryRepositories []string
	ArchiveGlobs      []string
	SessionID         string
	BucketMinutes     int
	DBPath            string
	TableName         string
}
