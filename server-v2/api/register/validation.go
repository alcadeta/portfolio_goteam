package register

import "regexp"

// Username defines the username field of a register request.
type Username string

// Validate applies username validation rules to the Username string and returns
// the error message if any fails.
func (u *Username) Validate() (errs []string) {
	if *u == "" {
		errs = append(errs, "Username cannot be empty.")
		// if username empty, further validation is pointless – return wantErrs
		return
	} else if len(*u) < 5 {
		errs = append(errs, "Username cannot be shorter than 5 characters.")
	} else if len(*u) > 15 {
		errs = append(errs, "Username cannot be longer than 15 characters.")
	}

	if match, _ := regexp.MatchString("[^A-Za-z0-9]+", string(*u)); match {
		errs = append(errs, "Username can contain only letters (a-z/A-Z) and digits (0-9).")
	}
	if match, _ := regexp.MatchString("(^\\d)", string(*u)); match {
		errs = append(errs, "Username can start only with a letter (a-z/A-Z).")
	}

	return
}

// Password defines the password field of a register request.
type Password string

// Validate applies password validation rules to the Password string and returns
// the error message if any fails.
func (p *Password) Validate() (errs []string) {
	if *p == "" {
		errs = append(errs, "Password cannot be empty.")
		// if password empty, further validation is pointless – return wantErrs
		return
	} else if len(*p) < 8 {
		errs = append(errs, "Password cannot be shorter than 8 characters.")
	} else if len(*p) > 64 {
		errs = append(errs, "Password cannot be longer than 64 characters.")
	}

	if match, _ := regexp.MatchString("[a-z]", string(*p)); !match {
		errs = append(errs, "Password must contain a lowercase letter (a-z).")
	}
	if match, _ := regexp.MatchString("[A-Z]", string(*p)); !match {
		errs = append(errs, "Password must contain an uppercase letter (A-Z).")
	}
	if match, _ := regexp.MatchString("[0-9]", string(*p)); !match {
		errs = append(errs, "Password must contain a digit (0-9).")
	}
	if match, _ := regexp.MatchString("[!\"#$%&'()*+,-./:;<=>?[\\]^_`{|}~]", string(*p)); !match {
		errs = append(errs, "Password must contain one of the following special characters: "+
			"! \" # $ % & ' ( ) * + , - . / : ; < = > ? [ \\ ] ^ _ ` { | } ~.")
	}
	if match, _ := regexp.MatchString("\\s", string(*p)); match {
		errs = append(errs, "Password cannot contain spaces.")
	}
	if match, _ := regexp.MatchString("[^\\x00-\\x7F]", string(*p)); match {
		errs = append(errs, "Password can contain only letters (a-z/A-Z), digits (0-9), "+
			"and the following special characters: "+
			"! \" # $ % & ' ( ) * + , - . / : ; < = > ? [ \\ ] ^ _ ` { | } ~.",
		)
	}

	return
}
