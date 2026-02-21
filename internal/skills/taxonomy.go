package skills

import "strings"

// Canonical skill names by category
var Skills = []string{
    // Languages
    "Go", "Python", "Java", "JavaScript", "TypeScript", "Rust",
    "C++", "C#", "Ruby", "PHP", "Kotlin", "Swift", "SQL", "R", "Scala",
    // Frameworks
    "React", "Next.js", "Vue", "Angular", "Django", "Flask",
    "Spring Boot", "Express", "FastAPI", "Gin", "Echo", "Fiber", "Rails",
    "Node.js",
    // Databases
    "PostgreSQL", "MySQL", "MongoDB", "Redis", "SQLite", "DynamoDB",
    "Cassandra", "Elasticsearch", "Neo4j", "ClickHouse", "Snowflake",
    // Cloud
    "AWS", "GCP", "Azure", "S3", "EC2", "Lambda", "Cloud Run",
    "BigQuery", "ECS", "EKS", "GKE", "CloudFormation",
    // DevOps
    "Docker", "Kubernetes", "Terraform", "Ansible", "Jenkins",
    "GitHub Actions", "CircleCI", "ArgoCD", "Helm", "Pulumi",
    // Tools & Protocols
    "Git", "Linux", "Nginx", "Kafka", "RabbitMQ", "gRPC", "GraphQL",
    "REST", "Prometheus", "Grafana", "Datadog", "OpenTelemetry",
    // Concepts
    "microservices", "CI/CD", "distributed systems", "system design",
    "API design", "event-driven", "caching", "load balancing",
    "message queue", "observability",
}

// Aliases maps alternate names/abbreviations to canonical names
var Aliases = map[string]string{
    "js":               "JavaScript",
    "ts":               "TypeScript",
    "k8s":              "Kubernetes",
    "kube":             "Kubernetes",
    "postgres":         "PostgreSQL",
    "pg":               "PostgreSQL",
    "mongo":            "MongoDB",
    "node":             "Node.js",
    "nodejs":           "Node.js",
    "gh actions":       "GitHub Actions",
    "google cloud":     "GCP",
    "microsoft azure":  "Azure",
    "rabbit":           "RabbitMQ",
    "rabbitmq":         "RabbitMQ",
    "elk":              "Elasticsearch",
    "elastic":          "Elasticsearch",
    "cicd":             "CI/CD",
    "grpc":             "gRPC",
    "rest api":         "REST",
    "restful":          "REST",
    "amazon web services": "AWS",
    "google kubernetes engine": "GKE",
    "amazon eks":       "EKS",
    "amazon ecs":       "ECS",
	"golang":      		"Go",
	"python3":     		"Python",
	"react.js":    		"React",
	"reactjs":     		"React",
	"vue.js":      		"Vue",
	"vuejs":       		"Vue",
	"nextjs":      		"Next.js",
	"angular.js":  		"Angular",
	"angularjs":   		"Angular",
	"typescript":  		"TypeScript",
	"javascript":  		"JavaScript",
	"postgresql":  		"PostgreSQL",
	"kubernetes":  		"Kubernetes",
	"prometheus":  		"Prometheus",
}

// index is a lowercase -> canonical map for O(1) lookup
var index map[string]string

func init() {
	index = make(map[string]string, len(Skills)+len(Aliases))
	for _, s := range Skills {
		index[strings.ToLower(s)] = s
 	}
	for alias, canonical := range Aliases {
		index[strings.ToLower(alias)] = canonical
	}
}

// Normalize returns the canonical name for a skill string.
// Returns the input as-is if not recognized.
func Normalize(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	if canonical, ok := index[lower]; ok {
		return canonical
	}
	return s
}

// IsKnown returns true if the skill is in the taxonomy.
func IsKnown(s string) bool {
	_, ok := index[strings.ToLower(strings.TrimSpace(s))]
	return ok
}