import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faGithub, faLinkedin } from '@fortawesome/free-brands-svg-icons'
import { useState } from "react"
import './App.css'

function App(){
  
  const [url, setUrl] = useState('')
  const [shortUrl, setShortUrl] = useState('')
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const generatedLink = shortUrl !== ''
  const [error, setError] = useState('')
  
  async function handleGenerate() {
    if (loading) return
  
    setError('')
    setLoading(true)

    try{
      const response = await fetch('/api/links', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url: url}),
      })

      if (!response.ok){
        const message = response.status === 429
          ? 'Calma lá amigão, tá rapidão ein? Limite de requisições atingido. Tente novamente em 10 minutos...'
          : 'Erro ao gerar o link. Verifique se a URL é válida.'
        setError(message)
        return
      }

      const data = await response.json()

      setShortUrl(`${window.location.origin}/${data.shortCode}`)
    }catch (error){
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy(){
    await navigator.clipboard.writeText(shortUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function handleReset(){
    setUrl('')
    setShortUrl('')
    setError('')
    setCopied(false)
  }

  return(
    <div className="container">
      <h1 className="title">LINKFORGE</h1>
      <div className="inputSearch">
        {generatedLink ? (
          <a
            className='input inputLink'
            href={shortUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            {shortUrl}
          </a>
        ) : (
          <input
            className="input"
            type="text"
            placeholder="Insira seu link aqui para forjar um novo 🔥🔨"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
          />
        )}
        {generatedLink ? (
          <button className="button" onClick={handleCopy}>
            {copied ? 'Copiado!' : 'Copiar'}
          </button>
        ) : (
          <button
          className="button"
          onClick={handleGenerate}
          disabled={!url || loading}
          >
            {loading ? '...' : 'Forjar Link!'}
          </button>
        )}
      </div>
      {generatedLink && (
        <button className="resetButton" onClick={handleReset}>
          Forjar outro Link
        </button>
      )}
      {error && <p className="errorMessage">{error}</p>}
      <div className="socials">
        <p className="socialsLabel">2026 - Desenvolvido por Leonardo Lima</p>

        <div className="socialsIcons">
          <a href="https://github.com/leonardolimaArt" target="_blank" className="socialButton" aria-label="Github">
            <FontAwesomeIcon icon={faGithub} size="lg"/>
          </a>
          <a href="https://www.linkedin.com/in/leonardolima-art" target="_blank" className="socialButton" aria-label="LinkedIn">
            <FontAwesomeIcon icon={faLinkedin} size="lg"/>
          </a>
        </div>
      </div>

    </div>
  )
}

export default App