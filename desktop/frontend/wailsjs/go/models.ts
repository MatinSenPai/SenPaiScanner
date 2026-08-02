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
	export class ScanParams {
	    count: number;
	    workers: number;
	    timeoutMs: number;
	    tries: number;
	    port: number;
	    mode: string;
	    sni: string;
	    speedTest: boolean;
	    requireWS: boolean;
	    coloFilter: string;
	    outputFile: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.workers = source["workers"];
	        this.timeoutMs = source["timeoutMs"];
	        this.tries = source["tries"];
	        this.port = source["port"];
	        this.mode = source["mode"];
	        this.sni = source["sni"];
	        this.speedTest = source["speedTest"];
	        this.requireWS = source["requireWS"];
	        this.coloFilter = source["coloFilter"];
	        this.outputFile = source["outputFile"];
	    }
	}
	export class ScanResult {
	    ip: string;
	    port: number;
	    colo: string;
	    avgMs: number;
	    minMs: number;
	    loss: number;
	    jitterMs: number;
	    throughput: number;
	    healthy: boolean;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.colo = source["colo"];
	        this.avgMs = source["avgMs"];
	        this.minMs = source["minMs"];
	        this.loss = source["loss"];
	        this.jitterMs = source["jitterMs"];
	        this.throughput = source["throughput"];
	        this.healthy = source["healthy"];
	        this.mode = source["mode"];
	    }
	}
	export class ValidationParams {
	    configUrl: string;
	    topN: number;
	    timeoutMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ValidationParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configUrl = source["configUrl"];
	        this.topN = source["topN"];
	        this.timeoutMs = source["timeoutMs"];
	    }
	}

}

