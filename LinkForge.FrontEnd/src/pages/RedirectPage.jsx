import { useLayoutEffect } from "react";
import { useParams } from "react-router-dom";

export default function RedirectPage(){
    const {shortCode} = useParams()
    useLayoutEffect(() => {
        const apiUrl = import.meta.env.VITE_API_URL || ''
        window.location.replace(`${apiUrl}/r/${shortCode}`)
    }, [shortCode])

    return null

}