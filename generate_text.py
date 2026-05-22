import os
import sys
import random
import time

def generate_vocab(size):
    vowels = "aeiou"
    consonants = "bcdfghjklmnpqrstvwxyz"
    vocab = []
    seen = set()
    while len(vocab) < size:
        length = random.randint(3, 9)
        word = "".join(random.choice(consonants if i % 2 == 0 else vowels) for i in range(length))
        if word not in seen:
            seen.add(word)
            vocab.append(word)
            
    # Pre-apply formatting to the vocabulary to speed up generation
    punctuations = [".", ",", "!", "?", ""]
    p_weights = [0.05, 0.05, 0.01, 0.01, 0.88]
    
    formatted_vocab = []
    for w in vocab:
        r = random.random()
        if r < 0.03:
            w = w.upper()
        elif r < 0.08:
            w = w.capitalize()
            
        p = random.choices(punctuations, weights=p_weights, k=1)[0]
        formatted_vocab.append(w + p)
        
    return formatted_vocab

def main():
    target_size_gb = 1.1
    target_size_bytes = int(target_size_gb * 1024 * 1024 * 1024)
    filename = "large_text.txt"
    vocab_size = 50000

    print(f"Generating vocabulary of {vocab_size} words...")
    vocab = generate_vocab(vocab_size)
    
    # Calculate Zipf-like weights
    print("Calculating weights...")
    weights = [1.0 / (i + 1)**0.8 for i in range(vocab_size)]
    
    print(f"Writing to {filename} (target: {target_size_gb} GB)...")
    start_time = time.time()
    written_bytes = 0
    
    # We will generate in chunks to speed up random.choices
    chunk_word_count = 250000
    
    with open(filename, "w", encoding="utf-8") as f:
        while written_bytes < target_size_bytes:
            words = random.choices(vocab, weights=weights, k=chunk_word_count)
            
            # Format chunk: join groups of 20 words with space, then join groups with newline
            sub_chunks = [" ".join(words[i:i+20]) for i in range(0, len(words), 20)]
            chunk_str = "\n".join(sub_chunks) + "\n"
            
            f.write(chunk_str)
            written_bytes += len(chunk_str.encode("utf-8"))
            
            elapsed = time.time() - start_time
            speed = written_bytes / (1024 * 1024 * elapsed) if elapsed > 0 else 0
            percent = (written_bytes / target_size_bytes) * 100
            print(f"\rProgress: {written_bytes / (1024*1024):.1f} MB / {target_size_gb*1024:.1f} MB ({percent:.1f}%) | Speed: {speed:.1f} MB/s", end="", flush=True)
            
    print(f"\nDone! File {filename} successfully generated in {time.time() - start_time:.1f} seconds.")

if __name__ == "__main__":
    main()
