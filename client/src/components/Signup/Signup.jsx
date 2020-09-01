import React, {useEffect, useState} from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import TextField from '@material-ui/core/TextField';
import Grid from '@material-ui/core/Grid';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import {Link, Redirect} from "react-router-dom";
import {List} from "@material-ui/core";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import axios from 'axios';
import {API_CONFIG} from "../../config/api";
import * as qs from "querystring";
import CircularProgress from "@material-ui/core/CircularProgress";
import {useUserContext} from "../../context/UserContext";
import Cookies from "universal-cookie";




// Template from: https://github.com/mui-org/material-ui/blob/master/docs/src/pages/getting-started/templates/sign-up/SignUp.js
const useStyles = makeStyles((theme) => ({
    paper: {
        marginTop: theme.spacing(8),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    avatar: {
        margin: theme.spacing(1),
        backgroundColor: theme.palette.secondary.main,
    },
    form: {
        width: '100%', // Fix IE 11 issue.
        marginTop: theme.spacing(3),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
    error: {
        color: 'red',
    },
    passwordOK: {
        color: '#066324'
    }
}));

const Signup = () => {
    const classes = useStyles();
    const cookies = new Cookies();

    const [user, dispatch] = useUserContext();

    useEffect(() => {
        if (!user.name) {
            const cookieUser = cookies.get('userinfo');
            if (cookieUser) {
                dispatch({type: 'setUser', user: cookieUser});
            }
        }
    }, [cookies, dispatch, user]);


    const [success, setSuccess] = useState(false);
    const [passwordFirstLoad, setPasswordFirstLoad] = useState(true);

    const [name, setName] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [email, setEmail] = useState("");

    const [nameError, setNameError] = useState(false);
    const [emailError, setEmailError] = useState(false);

    const [passwordsMatch, setPasswordsMatch] = useState(true);
    const [passwordCharLimitMet, setPasswordCharLimitMet] = useState(false);
    const [passwordUppercaseMet, setPasswordUppercaseMet] = useState(false);
    const [passwordLowercaseMet, setPasswordLowercaseMet] = useState(false);
    const [passwordDigitMet, setPasswordDigitMet] = useState(false);

    const [signingUp, setSigningUp] = useState(false);


    const [errorMessage, setErrorMessage] = useState("");

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!signingUp) {
            if (!validateNoErrors()) {
                setErrorMessage("One or more errors were found. Please check the form for details.")
            } else {
                setSigningUp(true);
                setErrorMessage("");

                const config = {
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded'
                    }
                }

                const body = {
                    name,
                    email,
                    password,
                }

                try {
                    await axios.post(`${API_CONFIG.base_url}/users/signup`, qs.stringify(body), config);
                    setSuccess(true);
                } catch (err) {
                    if (err.response && err.response.data) {
                        setErrorMessage(err.response.data);
                    } else {
                        setErrorMessage("Unknown error occurred while signing up; try again later")
                    }
                } finally {
                    setSigningUp(false);
                }
            }
        }

    }
    const handleNameChange = (e) => {
        setName(e.target.value);

        if (e.target.value === "") {
         setNameError(true);
        } else {
            setNameError(false);
        }
    }

    const validateNoErrors = () => {
        return !nameError && name !== '' && email !== '' && !emailError && passwordsMatch && passwordCharLimitMet && passwordUppercaseMet && passwordDigitMet
    }

    const handleEmailChange = (e) => {
        setEmail(e.target.value);
        const regex = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
        if (e.target.value === "" || !regex.test(e.target.value)) {
            setEmailError(true);
        } else {
            setEmailError(false);
        }
    }
    const handlePasswordChange = (event, confirm) => {
        setPasswordFirstLoad(false);
        if (confirm) {
            setConfirmPassword(event.target.value);
        } else {
            setPassword(event.target.value);
        }
    }

    useEffect(() => {
        setPasswordsMatch(password === confirmPassword);
        setPasswordCharLimitMet(password.length >= 8);
        let upperCase = false;
        let lowerCase = false;
        let digit = false;

        for (let i = 0; i < password.length; i++) {
            const cur = password.charAt(i);

            if (cur >= '0' && cur <= '9') digit = true;
            else if (cur >= 'a' && cur <= 'z') lowerCase = true;
            else if (cur >= 'A' && cur <= 'Z') upperCase = true;
        }

        setPasswordUppercaseMet(upperCase);
        setPasswordLowercaseMet(lowerCase);
        setPasswordDigitMet(digit);
    }, [password, confirmPassword])

    if (user.name) {
        return (
            <Redirect to={"/"}/>
        )
    }

    if (success) {
        return (
            <Redirect to={"/login?origin=signup"} />
        )
    }

    return (
        <Container component="main" maxWidth="xs">
            <CssBaseline />
            <div className={classes.paper}>
                <Avatar className={classes.avatar}>
                    <LockOutlinedIcon />
                </Avatar>
                <Typography component="h1" variant="h5">
                    Sign up
                </Typography>
                <form className={classes.form} noValidate onSubmit={handleSubmit}>
                    <Grid container spacing={2}>
                        <Grid item xs={12}>
                            <TextField
                                value={name}
                                autoComplete="name"
                                name="name"
                                variant="outlined"
                                helperText={(nameError) ? "Please enter a name." : ""}
                                required
                                fullWidth
                                id="name"
                                label="Name"
                                autoFocus
                                error={nameError}
                                onChange={handleNameChange}
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <TextField
                                value={email}
                                variant="outlined"
                                helperText={(emailError) ? "Please enter a valid email address." : ""}
                                required
                                fullWidth
                                id="email"
                                label="Email Address"
                                name="email"
                                autoComplete="email"
                                error={emailError}
                                onChange={handleEmailChange}
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <TextField
                                value={password}
                                variant="outlined"
                                required
                                fullWidth
                                name="password"
                                label="Password"
                                type="password"
                                id="password"
                                autoComplete="current-password"
                                error={!passwordFirstLoad && (!passwordCharLimitMet || !passwordLowercaseMet || !passwordUppercaseMet || !passwordDigitMet) }
                                onChange={(e) =>
                                    handlePasswordChange(e, false)
                                }
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <TextField
                                value={confirmPassword}
                                variant="outlined"
                                required
                                fullWidth
                                name="confirm-password"
                                label="Confirm Password"
                                type="password"
                                id="confirm-password"
                                autoComplete="current-password"
                                error={!passwordsMatch}
                                onChange={(e) =>
                                    handlePasswordChange(e, true)
                                }
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <Typography variant="h6" className={classes.title}>
                                Password complexity requirements:
                            </Typography>
                            <List dense={true}>
                                <ListItem><ListItemText className={(passwordCharLimitMet) ? classes.passwordOK : classes.error} primary={"Has at least 8 characters"} /></ListItem>
                                <ListItem><ListItemText className={(passwordUppercaseMet) ? classes.passwordOK : classes.error} primary={"Has at least 1 UPPERCASE letter"}/></ListItem>
                                <ListItem><ListItemText className={(passwordLowercaseMet) ? classes.passwordOK: classes.error} primary={"Has at least 1 lowercase letter"} /></ListItem>
                                <ListItem><ListItemText className={(passwordDigitMet) ? classes.passwordOK : classes.error} primary={"Has at least 1 number"} /></ListItem>
                                <ListItem><ListItemText className={(passwordsMatch) ? classes.passwordOK : classes.error} primary={"Password and confirm password fields must match"} /></ListItem>
                            </List>
                        </Grid>
                    </Grid>
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        disabled={!validateNoErrors()}
                    >
                        {signingUp ? <CircularProgress color={"secondary"}/> : "Sign Up"}
                    </Button>
                    <Grid container justify="flex-end">
                        <Grid item>
                            <Link to="/login" replace>
                                Already have an account? Log in
                            </Link>
                        </Grid>
                    </Grid>
                    <p className={classes.error}>{errorMessage}</p>
                </form>
            </div>
        </Container>
    );
}

export default Signup;