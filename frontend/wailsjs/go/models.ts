export namespace downloader {
	
	export class Config {
	    maxConcurrentDownloads: number;
	    maxPartsPerDownload: number;
	    downloadDir: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxConcurrentDownloads = source["maxConcurrentDownloads"];
	        this.maxPartsPerDownload = source["maxPartsPerDownload"];
	        this.downloadDir = source["downloadDir"];
	    }
	}
	export class Download {
	    id: string;
	    url: string;
	    fileName: string;
	    totalSize: number;
	    downloaded: number;
	    speed: number;
	    status: string;
	    parts: number;
	    error?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Download(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.fileName = source["fileName"];
	        this.totalSize = source["totalSize"];
	        this.downloaded = source["downloaded"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.parts = source["parts"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class DownloadInfo {
	    url: string;
	    fileName: string;
	    totalSize: number;
	    contentType: string;
	    resumable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DownloadInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.fileName = source["fileName"];
	        this.totalSize = source["totalSize"];
	        this.contentType = source["contentType"];
	        this.resumable = source["resumable"];
	    }
	}

}

