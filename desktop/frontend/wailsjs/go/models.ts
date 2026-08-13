export namespace main {
	
	export class ExportBundle {
	    subscription: string;
	    shareUrls: string[];
	    singBox: string;
	    clash: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ExportBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subscription = source["subscription"];
	        this.shareUrls = source["shareUrls"];
	        this.singBox = source["singBox"];
	        this.clash = source["clash"];
	        this.count = source["count"];
	    }
	}
	export class PresetData {
	    countLabels: string[];
	    countValues: string[];
	    workerLabels: string[];
	    workerValues: string[];
	    timeoutLabels: string[];
	    timeoutValues: string[];
	    topNLabels: string[];
	    topNValues: string[];
	    minSpeedLabels: string[];
	    minSpeedValues: string[];
	    speedSizeLabels: string[];
	    speedSizeValues: string[];
	    ports: number[];
	
	    static createFrom(source: any = {}) {
	        return new PresetData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.countLabels = source["countLabels"];
	        this.countValues = source["countValues"];
	        this.workerLabels = source["workerLabels"];
	        this.workerValues = source["workerValues"];
	        this.timeoutLabels = source["timeoutLabels"];
	        this.timeoutValues = source["timeoutValues"];
	        this.topNLabels = source["topNLabels"];
	        this.topNValues = source["topNValues"];
	        this.minSpeedLabels = source["minSpeedLabels"];
	        this.minSpeedValues = source["minSpeedValues"];
	        this.speedSizeLabels = source["speedSizeLabels"];
	        this.speedSizeValues = source["speedSizeValues"];
	        this.ports = source["ports"];
	    }
	}
	export class ScanParams {
	    ipMode: number;
	    ipFile: string;
	    count: number;
	    workers: number;
	    timeoutMs: number;
	    ports: number[];
	    configUrl: string;
	    requireWS: boolean;
	    topN: number;
	    minSpeed: number;
	    speedUrl: string;
	    speedSize: number;
	    uploadTest: boolean;
	    uploadSize: number;
	    uploadUrl: string;
	    neighborScan: boolean;
	    countIdx: number;
	    countCustom: string;
	    workersIdx: number;
	    workersCustom: string;
	    timeoutIdx: number;
	    timeoutCustom: string;
	    topNIdx: number;
	    topNCustom: string;
	    minSpeedIdx: number;
	    minSpeedCustom: string;
	    speedSizeIdx: number;
	    speedSizeCustom: string;
	    uploadSizeIdx: number;
	    uploadSizeCustom: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ipMode = source["ipMode"];
	        this.ipFile = source["ipFile"];
	        this.count = source["count"];
	        this.workers = source["workers"];
	        this.timeoutMs = source["timeoutMs"];
	        this.ports = source["ports"];
	        this.configUrl = source["configUrl"];
	        this.requireWS = source["requireWS"];
	        this.topN = source["topN"];
	        this.minSpeed = source["minSpeed"];
	        this.speedUrl = source["speedUrl"];
	        this.speedSize = source["speedSize"];
	        this.uploadTest = source["uploadTest"];
	        this.uploadSize = source["uploadSize"];
	        this.uploadUrl = source["uploadUrl"];
	        this.neighborScan = source["neighborScan"];
	        this.countIdx = source["countIdx"];
	        this.countCustom = source["countCustom"];
	        this.workersIdx = source["workersIdx"];
	        this.workersCustom = source["workersCustom"];
	        this.timeoutIdx = source["timeoutIdx"];
	        this.timeoutCustom = source["timeoutCustom"];
	        this.topNIdx = source["topNIdx"];
	        this.topNCustom = source["topNCustom"];
	        this.minSpeedIdx = source["minSpeedIdx"];
	        this.minSpeedCustom = source["minSpeedCustom"];
	        this.speedSizeIdx = source["speedSizeIdx"];
	        this.speedSizeCustom = source["speedSizeCustom"];
	        this.uploadSizeIdx = source["uploadSizeIdx"];
	        this.uploadSizeCustom = source["uploadSizeCustom"];
	    }
	}
	export class ScanResult {
	    ip: string;
	    port: number;
	    colo: string;
	    avgMs: number;
	    loss: number;
	    jitterMs: number;
	    throughput: number;
	    healthy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.colo = source["colo"];
	        this.avgMs = source["avgMs"];
	        this.loss = source["loss"];
	        this.jitterMs = source["jitterMs"];
	        this.throughput = source["throughput"];
	        this.healthy = source["healthy"];
	    }
	}
	export class ValidationOutcome {
	    ip: string;
	    port: number;
	    transport: string;
	    success: boolean;
	    latencyMs: number;
	    throughput: number;
	    uploadThroughput: number;
	    error: string;
	    done: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ValidationOutcome(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.transport = source["transport"];
	        this.success = source["success"];
	        this.latencyMs = source["latencyMs"];
	        this.throughput = source["throughput"];
	        this.uploadThroughput = source["uploadThroughput"];
	        this.error = source["error"];
	        this.done = source["done"];
	        this.total = source["total"];
	    }
	}

}

