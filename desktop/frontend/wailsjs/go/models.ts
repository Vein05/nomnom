export namespace main {
	
	export class VisionConfig {
	    enabled: boolean;
	    max_image_size: string;
	
	    static createFrom(source: any = {}) {
	        return new VisionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.max_image_size = source["max_image_size"];
	    }
	}
	export class AIConfig {
	    provider: string;
	    model: string;
	    api_key?: string;
	    max_tokens: number;
	    temperature: number;
	    vision: VisionConfig;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.api_key = source["api_key"];
	        this.max_tokens = source["max_tokens"];
	        this.temperature = source["temperature"];
	        this.vision = this.convertValues(source["vision"], VisionConfig);
	        this.prompt = source["prompt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OllamaStatus {
	    running: boolean;
	    model_available: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new OllamaStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.model_available = source["model_available"];
	        this.message = source["message"];
	    }
	}
	export class AIStatus {
	    provider: string;
	    model: string;
	    ollama: OllamaStatus;
	
	    static createFrom(source: any = {}) {
	        return new AIStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.ollama = this.convertValues(source["ollama"], OllamaStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AnalyticsSummary {
	    sessions: number;
	    renamed: number;
	    tokens: number;
	    avg_per_run: number;
	    recent_runs: number;
	    unique_models: number;
	    history_error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalyticsSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessions = source["sessions"];
	        this.renamed = source["renamed"];
	        this.tokens = source["tokens"];
	        this.avg_per_run = source["avg_per_run"];
	        this.recent_runs = source["recent_runs"];
	        this.unique_models = source["unique_models"];
	        this.history_error = source["history_error"];
	    }
	}
	export class ContentExtractionConfig {
	    extract_text: boolean;
	    extract_metadata: boolean;
	    max_content_length: number;
	    skip_large_files: boolean;
	    read_context: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContentExtractionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.extract_text = source["extract_text"];
	        this.extract_metadata = source["extract_metadata"];
	        this.max_content_length = source["max_content_length"];
	        this.skip_large_files = source["skip_large_files"];
	        this.read_context = source["read_context"];
	    }
	}
	export class LoggingConfig {
	    enabled: boolean;
	    log_path: string;
	
	    static createFrom(source: any = {}) {
	        return new LoggingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.log_path = source["log_path"];
	    }
	}
	export class PerformancePipelineConfig {
	    workers: number;
	    timeout: string;
	    retries: number;
	
	    static createFrom(source: any = {}) {
	        return new PerformancePipelineConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workers = source["workers"];
	        this.timeout = source["timeout"];
	        this.retries = source["retries"];
	    }
	}
	export class PerformanceConfig {
	    ai: PerformancePipelineConfig;
	    file: PerformancePipelineConfig;
	
	    static createFrom(source: any = {}) {
	        return new PerformanceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ai = this.convertValues(source["ai"], PerformancePipelineConfig);
	        this.file = this.convertValues(source["file"], PerformancePipelineConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileHandlingConfig {
	    max_size: string;
	    auto_approve: boolean;
	    hot_rename: boolean;
	    skip_dot_files: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileHandlingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.max_size = source["max_size"];
	        this.auto_approve = source["auto_approve"];
	        this.hot_rename = source["hot_rename"];
	        this.skip_dot_files = source["skip_dot_files"];
	    }
	}
	export class DesktopConfig {
	    output: string;
	    case: string;
	    ai: AIConfig;
	    file_handling: FileHandlingConfig;
	    content_extraction: ContentExtractionConfig;
	    performance: PerformanceConfig;
	    logging: LoggingConfig;
	
	    static createFrom(source: any = {}) {
	        return new DesktopConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	        this.case = source["case"];
	        this.ai = this.convertValues(source["ai"], AIConfig);
	        this.file_handling = this.convertValues(source["file_handling"], FileHandlingConfig);
	        this.content_extraction = this.convertValues(source["content_extraction"], ContentExtractionConfig);
	        this.performance = this.convertValues(source["performance"], PerformanceConfig);
	        this.logging = this.convertValues(source["logging"], LoggingConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class JobSummary {
	    planned: number;
	    renamed: number;
	    skipped: number;
	    errors: number;
	
	    static createFrom(source: any = {}) {
	        return new JobSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.planned = source["planned"];
	        this.renamed = source["renamed"];
	        this.skipped = source["skipped"];
	        this.errors = source["errors"];
	    }
	}
	export class JobStatus {
	    job_id: string;
	    state: string;
	    done: number;
	    total: number;
	    current_file: string;
	    message: string;
	    summary: JobSummary;
	    output_dir: string;
	
	    static createFrom(source: any = {}) {
	        return new JobStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.job_id = source["job_id"];
	        this.state = source["state"];
	        this.done = source["done"];
	        this.total = source["total"];
	        this.current_file = source["current_file"];
	        this.message = source["message"];
	        this.summary = this.convertValues(source["summary"], JobSummary);
	        this.output_dir = source["output_dir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class OpenRouterKeyStatus {
	    available: boolean;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenRouterKeyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.source = source["source"];
	    }
	}
	export class OpenRouterTestResult {
	    ok: boolean;
	    status_code: number;
	    status_text: string;
	    source: string;
	    message: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenRouterTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.status_code = source["status_code"];
	        this.status_text = source["status_text"];
	        this.source = source["source"];
	        this.message = source["message"];
	        this.response = source["response"];
	    }
	}
	
	
	export class RenameEntry {
	    index: number;
	    original: string;
	    new_name: string;
	    type: string;
	    status: string;
	    size_bytes: number;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new RenameEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.original = source["original"];
	        this.new_name = source["new_name"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.size_bytes = source["size_bytes"];
	        this.reason = source["reason"];
	    }
	}
	export class RunJobOptions {
	    dry_run: boolean;
	    log_session: boolean;
	    auto_approve: boolean;
	    hot_rename: boolean;
	    organize: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunJobOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dry_run = source["dry_run"];
	        this.log_session = source["log_session"];
	        this.auto_approve = source["auto_approve"];
	        this.hot_rename = source["hot_rename"];
	        this.organize = source["organize"];
	    }
	}
	export class Session {
	    date: string;
	    directory: string;
	    files: string;
	    model: string;
	    mode: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.directory = source["directory"];
	        this.files = source["files"];
	        this.model = source["model"];
	        this.mode = source["mode"];
	        this.status = source["status"];
	    }
	}

}

