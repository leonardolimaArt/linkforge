import { useState } from "react"
import './App.css'

function App(){
  
  const [url, setUrl] = useState('')
  const [shortUrl, setShortUrl] = useState('')
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const generatedLink = shortUrl !== ''
  
  async function handleGenerate() {
    if (loading) return
  
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
        throw new Error('Erro ao gerar o link. Verifique se a URL é válida.')
      }

      const data = await response.json()

      setShortUrl(`${window.location.origin}/${data.shortCode}`)
    }catch (error){
      alert(error.message)
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
    setCopied(false)
  }

  return(
    <div className="container">
      <h1 className="title">LINKFORGE</h1>
      <div className="inputSearch">
        <input
          className="input"
          type="text"
          placeholder="Insira com força o seu link aqui"
          value={generatedLink ? shortUrl : url}
          onChange={(e) => setUrl(e.target.value)}
          readOnly={generatedLink}
        />
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
            {loading ? '...' : 'Gerar Link!'}
          </button>
        )}
      </div>
      {generatedLink && (
        <button className="resetButton" onClick={handleReset}>
          Gerar outro Link
        </button>
      )}
    </div>
  )
}

export default App