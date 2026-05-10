module eino-intern-workflows

go 1.22

require (
	github.com/cloudwego/eino v0.8.0
)

// Notes:
// 1. This project is an interview-grade reconstruction of the internship workflows.
// 2. Milvus/ES/LLM clients are injected through interfaces, so you can swap mock clients
//    with real clients in production without changing the workflow topology.
