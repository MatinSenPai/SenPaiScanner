export namespace antidpi {
	
	export class Profile {
	    enabled: boolean;
	    finalmask: string;
	    fingerprint: string;
	    alpn: string;
	    cipherSuites: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.finalmask = source["finalmask"];
	        this.fingerprint = source["fingerprint"];
	        this.alpn = source["alpn"];
	        this.cipherSuites = source["cipherSuites"];
	    }
	}

}

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
	    neighborScan: boolean;
	    antiDpi: antidpi.Profile;
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
	
	    static createFrom(source: any = {}) {
	        return new ScanParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ipMode = source["ipMode"];
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
	        this.neighborScan = source["neighborScan"];
	        this.antiDpi = this.convertValues(source["antiDpi"], antidpi.Profile);
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

}

