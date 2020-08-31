import React, { useState } from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import TextField from '@material-ui/core/TextField';
import Grid from '@material-ui/core/Grid';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import {API_CONFIG} from "../../config/api";
import CircularProgress from "@material-ui/core/CircularProgress";
import * as qs from "querystring";
import axios from 'axios';
import {Link, Redirect} from "react-router-dom";
import {useUserContext} from "../../context/UserContext";
import {Cookies} from "react-cookie";

// Template from: https://material-ui.com/getting-started/templates/sign-in/

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
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
    error: {
        color: 'red'
    }
}));

const Login = () => {
    const cookies = new Cookies();

    const [user, dispatch] = useUserContext();

    const [loggingIn, setLoggingIn] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");

    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    const [emailError, setEmailError] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [success, setSuccess] = useState(false);

    const handleEmailChange = (event) => {
        setEmail(event.target.value);
        if (event.target.value === '') {
          setEmailError(true);
        } else {
            setEmailError(false);
        }
    }

    const handlePasswordChange = (event) => {
        setPassword(event.target.value);

        if (event.target.value === '') {
            setPasswordError(true);
        } else {
            setPasswordError(false);
        }
    }


    const handleSubmit = async (event) => {
        event.preventDefault();

        if (password === "" || email === "") {
            if (password === "" && email === "") {
                setPasswordError(true);
                setEmailError(true);
                setErrorMessage("Enter an email and password.")
            } else if (password === "") {
                setPasswordError(true);
                setErrorMessage("Enter a password.")
            } else {
                setEmailError(true);
                setErrorMessage("Enter an email.")
            }
        } else if (!loggingIn) {
            setErrorMessage("");
            setLoggingIn(true);

            const config = {
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                }
            }

            const body = {
                email,
                password
            }

            try {
                axios.defaults.withCredentials = true;
                const response = await axios.post(`${API_CONFIG.base_url}/users/login`, qs.stringify(body), config);

                const { token, expiry, user} = response.data;

                const expiryDateParsed = new Date(Date.parse(expiry));

                cookies.set("logintoken", token, {path: "/", expires: expiryDateParsed, httpOnly: true, secure: (process.env.NODE_ENV === 'production'), sameSite: "lax"});
                cookies.set("userinfo", user, { path: "/", expires: expiryDateParsed, httpOnly: false, secure: false, sameSite: "none"});

                dispatch({ type: 'setUser', user});
                setLoggingIn(false);
                setSuccess(true);
            } catch (err) {
                console.log(err);
                if (err.response && err.response.data) {
                    setErrorMessage(err.response.data);
                } else {
                    setErrorMessage("Unknown error occurred while logging in; try again later")
                }
            } finally {
                setLoggingIn(false);
            }
        }

    }

    const classes = useStyles();

    if (user.name || success) {
        return (
            <Redirect to={"/"} />
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
                    Sign in
                </Typography>
                <form className={classes.form} noValidate onSubmit={handleSubmit}>
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="email"
                        label="Email Address"
                        name="email"
                        autoComplete="email"
                        disabled={loggingIn}
                        value={email}
                        error={emailError}
                        onChange={handleEmailChange}
                        autoFocus
                    />
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        name="password"
                        label="Password"
                        type="password"
                        id="password"
                        error={passwordError}
                        disabled={loggingIn}
                        autoComplete="current-password"
                        value={password}
                        onChange={handlePasswordChange}
                    />
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        disabled={loggingIn}
                    >
                        {loggingIn ? <CircularProgress color={"secondary"}/> : "Sign In"}
                    </Button>
                    <Grid container>
                        <Grid item>
                            <Link to="/signup">
                                {"Don't have an account? Sign Up"}
                            </Link>
                        </Grid>
                    </Grid>
                    <p className={classes.error}>{errorMessage}</p>
                </form>
            </div>
        </Container>);
}

export default Login;