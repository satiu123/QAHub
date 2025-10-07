export namespace services {
	
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

