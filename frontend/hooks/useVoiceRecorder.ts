"use client";

import { useState, useRef, useEffect } from "react";
import { supabase } from "@/lib/supabase";

interface VoiceRecorderCallbacks {
    onTranscriptionStart?: () => void;
    onTranscriptionSuccess?: () => void;
    onTranscriptionError?: () => void;
}

export function useVoiceRecorder(
    onTranscriptionComplete: (text: string) => void,
    callbacks?: VoiceRecorderCallbacks
) {
    const [isRecording, setIsRecording] = useState(false);
    const [isTranscribing, setIsTranscribing] = useState(false);
    const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);
    const streamRef = useRef<MediaStream | null>(null);

    // Cleanup stream on unmount
    useEffect(() => {
        return () => {
            if (streamRef.current) {
                streamRef.current.getTracks().forEach(track => track.stop());
            }
        };
    }, []);

    const startRecording = async () => {
        try {
            let stream = streamRef.current;
            
            // Re-use existing stream if available and active, otherwise request new permission
            if (!stream || !stream.active) {
                stream = await navigator.mediaDevices.getUserMedia({ audio: true });
                streamRef.current = stream;
            }

            const recorder = new MediaRecorder(stream);
            const chunks: Blob[] = [];

            recorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    chunks.push(e.data);
                }
            };

            recorder.onstop = async () => {
                setIsTranscribing(true);
                callbacks?.onTranscriptionStart?.();
                
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
                    callbacks?.onTranscriptionSuccess?.();
                } catch (err) {
                    console.error('Transcription error:', err);
                    callbacks?.onTranscriptionError?.();
                    alert('Failed to transcribe audio. Please try again.');
                } finally {
                    setIsTranscribing(false);
                    // NOTE: We do NOT stop the stream tracks here.
                    // keeping the stream active allows subsequent recordings without re-prompting.
                }
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
