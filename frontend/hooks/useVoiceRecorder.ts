"use client";

import { useState } from "react";
import { supabase } from "@/lib/supabase";

export function useVoiceRecorder(onTranscriptionComplete: (text: string) => void) {
    const [isRecording, setIsRecording] = useState(false);
    const [isTranscribing, setIsTranscribing] = useState(false);
    const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);

    const startRecording = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const recorder = new MediaRecorder(stream);
            const chunks: Blob[] = [];

            recorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    chunks.push(e.data);
                }
            };

            recorder.onstop = async () => {
                setIsTranscribing(true);
                const audioBlob = new Blob(chunks, { type: 'audio/webm' });
                
                // Send to backend for transcription
                const formData = new FormData();
                formData.append('audio', audioBlob, 'recording.webm');

                try {
                    const { data: { session } } = await supabase.auth.getSession();
                    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/transcribe`, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${session?.access_token}`,
                        },
                        body: formData,
                    });

                    if (!response.ok) {
                        throw new Error('Transcription failed');
                    }

                    const result = await response.json();
                    onTranscriptionComplete(result.text);
                } catch (err) {
                    console.error('Transcription error:', err);
                    alert('Failed to transcribe audio. Please try again.');
                } finally {
                    setIsTranscribing(false);
                }

                // Stop all tracks
                stream.getTracks().forEach(track => track.stop());
            };

            recorder.start();
            setMediaRecorder(recorder);
            setIsRecording(true);
        } catch (err) {
            console.error('Failed to start recording:', err);
            alert('Failed to access microphone. Please check permissions.');
        }
    };

    const stopRecording = () => {
        if (mediaRecorder && mediaRecorder.state !== 'inactive') {
            mediaRecorder.stop();
            setIsRecording(false);
            setMediaRecorder(null);
        }
    };

    return {
        isRecording,
        isTranscribing,
        startRecording,
        stopRecording,
    };
}
