package bookinfo


// CommonLanguages — карта распространенных языков программирования.
// Ключ: алиас (в нижнем регистре), Значение: каноническое имя языка.
var CommonLanguages = map[string]string{
	// Backend / Systems
	"go":     "Go",
	"golang": "Go",
	"rs":     "Rust",
	"rust":   "Rust",
	"java":   "Java",
	"jvm":    "Java",
	"kt":     "Kotlin",
	"kotlin": "Kotlin",
	"scala":  "Scala",
	"c":      "C",
	"cpp":    "C++",
	"c++":    "C++",
	"cs":     "C#",
	"csharp": "C#",
	"net":    ".NET",
	"dotnet": ".NET",
	"elixir": "Elixir",
	"erlang": "Erlang",

	// Scripting / Web
	"py":     "Python",
	"python": "Python",
	"rb":     "Ruby",
	"ruby":   "Ruby",
	"php":    "PHP",
	"js":     "JavaScript",
	"javascript": "JavaScript",
	"ts":     "TypeScript",
	"perl":   "Perl",
	"lua":    "Lua",
	"bash":   "Bash",
	"sh":     "Bash",
	"shell":  "Bash",

	// Mobile
	"swift": "Swift",
	"dart":  "Dart",
	"obj-c": "Objective-C",

	// Data
	"sql": "SQL",
	"r":   "R",
}

// TopicTags — карта для определения категории (роли/сферы) по ключевому слову.
// Сюда входят фреймворки, инструменты и понятия.
// Ключ: алиас, Значение: канонический Тэг (Role/Category).
var TopicTags = map[string]string{
	// Frontend
	"html":    "Frontend",
	"css":     "Frontend",
	"scss":    "Frontend",
	"vue":     "Frontend",
	"react":   "Frontend",
	"angular": "Frontend",
	"webpack": "Frontend",
	"jquery":  "Frontend",
	"wasm":    "Frontend",

	// Backend
	"spring":        "Backend",
	"django":        "Backend",
	"flask":         "Backend",
	"fastapi":       "Backend",
	"laravel":       "Backend",
	"symfony":       "Backend",
	"rails":         "Backend",
	"node":          "Backend",
	"nodejs":        "Backend",
	"microservices": "Backend",
	"api":           "Backend",
	"grpc":          "Backend",

	// Mobile
	"ios":          "Mobile",
	"android":      "Mobile",
	"flutter":      "Mobile",
	"react-native": "Mobile",
	"swiftui":      "Mobile",

	// DevOps / Cloud
	"docker":     "DevOps",
	"k8s":        "DevOps",
	"kubernetes": "DevOps",
	"terraform":  "DevOps",
	"ansible":    "DevOps",
	"jenkins":    "DevOps",
	"gitlab":     "DevOps",
	"aws":        "DevOps",
	"linux":      "DevOps",
	"nginx":      "DevOps",
	"ci/cd":      "DevOps",

	// Data / ML
	"mongo":      "Data",
	"redis":      "Data",
	"postgres":   "Data",
	"mysql":      "Data",
	"kafka":      "Data",
	"pandas":     "Data Science",
	"numpy":      "Data Science",
	"tensorflow": "Machine Learning",
	"pytorch":    "Machine Learning",
	"ml":         "Machine Learning",
	"ai":         "AI",

	// Architecture / CS
	"algo":           "Computer Science",
	"algorithms":     "Computer Science",
	"patterns":       "Architecture",
	"architecture":   "Architecture",
	"system-design":  "Architecture",
	"distributed":    "Architecture",
	"clean-code":     "Best Practices",
	"refactoring":    "Best Practices",
	"tdd":            "Testing",
	"testing":        "Testing",

	// Security
	"security": "Security",
	"hacking":  "Security",
	"pentest":  "Security",
	"crypto":   "Security",
}

// LanguageDefaultCategories — вспомогательная карта.
// Если мы нашли Язык, но не нашли Тэг, мы можем присвоить дефолтный Тэг.
var LanguageDefaultCategories = map[string]string{
	"Go":         "Backend",
	"Java":       "Backend",
	"Kotlin":     "Mobile", // или Backend, зависит от контекста, но чаще сейчас Mobile
	"C#":         "Backend",
	"JavaScript": "Frontend",
	"TypeScript": "Frontend",
	"Python":     "Backend", // или Data Science, тут 50/50
	"PHP":        "Backend",
	"Swift":      "Mobile",
	"Dart":       "Mobile",
	"Rust":       "Backend",
	"C++":        "System",
	"C":          "System",
	"SQL":        "Data",
}
