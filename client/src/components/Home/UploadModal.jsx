
import React, {useEffect, useState} from 'react';
import Dialog from "@material-ui/core/Dialog";
import {Typography} from "@material-ui/core";
import Button from "@material-ui/core/Button";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import {makeStyles} from "@material-ui/core/styles";

const UploadModal = ({open, handleClose}) => {
    const fileInput = React.createRef();
    const [errorText, setErrorText] = useState("");
    const [file, setFile] = useState(null);

    const handleSubmit = (e) => {
        e.preventDefault();
    }

    const handleSelectImage = (e) => {
        const file = fileInput.current.files[0];
        if (file) {
            if (file.size <= 9000000) {
                setFile(fileInput.current.files[0]);
            } else {
                setFile(null);
                setErrorText("Selected image exceeds size limit (9MB)")
            }
        }
    }

    return (
        <Dialog open={open} onClose={() => {setFile(null); setErrorText(""); handleClose()}} fullWidth={true}>
            <DialogContent>
                <DialogContentText>Upload image</DialogContentText>
                <form onSubmit={handleSubmit} noValidate>
                    {!file ?
                        (<>
                            <input
                                accept="image/*"
                                style={{ display: 'none' }}
                                id="image-upload"
                                type="file"
                                ref={fileInput}
                                onChange={handleSelectImage}
                            />
                            <label htmlFor="image-upload">
                                <Button variant="contained" component="span">
                                    Upload
                                </Button>
                            </label>
                        </>) :
                        <Button variant="contained" component="span" onClick={() => setFile(null)}>
                            Delete
                        </Button>}
                </form>




            </DialogContent>

        </Dialog>
    )
}

export default UploadModal;