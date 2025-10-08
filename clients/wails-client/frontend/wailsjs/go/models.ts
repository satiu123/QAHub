export namespace main {
	
	export class NotificationListResult {
	    notifications: services.Notification[];
	    total: number;
	    unread_count: number;
	
	    static createFrom(source: any = {}) {
	        return new NotificationListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.notifications = this.convertValues(source["notifications"], services.Notification);
	        this.total = source["total"];
	        this.unread_count = source["unread_count"];
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

export namespace services {
	
	export class Answer {
	    id: number;
	    question_id: number;
	    content: string;
	    user_id: number;
	    username: string;
	    upvote_count: number;
	    is_upvoted: boolean;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Answer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.question_id = source["question_id"];
	        this.content = source["content"];
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.upvote_count = source["upvote_count"];
	        this.is_upvoted = source["is_upvoted"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class Comment {
	    id: number;
	    answer_id: number;
	    user_id: number;
	    username: string;
	    content: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Comment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.answer_id = source["answer_id"];
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.content = source["content"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class LoginResponse {
	    success: boolean;
	    token: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.token = source["token"];
	        this.message = source["message"];
	    }
	}
	export class Notification {
	    id: string;
	    recipient_id: number;
	    sender_id: number;
	    sender_name: string;
	    type: string;
	    content: string;
	    target_url: string;
	    is_read: boolean;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Notification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.recipient_id = source["recipient_id"];
	        this.sender_id = source["sender_id"];
	        this.sender_name = source["sender_name"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.target_url = source["target_url"];
	        this.is_read = source["is_read"];
	        this.created_at = source["created_at"];
	    }
	}
	export class Question {
	    id: number;
	    title: string;
	    content: string;
	    user_id: number;
	    author_name: string;
	    answer_count: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Question(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.user_id = source["user_id"];
	        this.author_name = source["author_name"];
	        this.answer_count = source["answer_count"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class RegisterResponse {
	    success: boolean;
	    user_id: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.user_id = source["user_id"];
	        this.message = source["message"];
	    }
	}
	export class SearchResult {
	    id: number;
	    title: string;
	    content: string;
	    author_id: number;
	    author_name: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.author_id = source["author_id"];
	        this.author_name = source["author_name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class UserProfile {
	    user_id: number;
	    username: string;
	    email: string;
	    bio: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new UserProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.bio = source["bio"];
	        this.created_at = source["created_at"];
	    }
	}

}

