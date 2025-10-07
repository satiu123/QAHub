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
	export class UserProfile {
	    user_id: number;
	    username: string;
	    email: string;
	    bio: string;
	
	    static createFrom(source: any = {}) {
	        return new UserProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.bio = source["bio"];
	    }
	}

}

